package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/constants"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/db"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

func (s *JobService) processRenderSubmission(ctx context.Context, job models.Job) error {
	job.Metadata = ensureJobMetadata(job.Metadata)
	if requestID := metadataString(job.Metadata, internalRenderRequestIDKey); requestID != "" {
		return nil
	}
	if job.SelectedSlotID == nil || *job.SelectedSlotID == "" {
		return s.failRenderJob(ctx, job, models.Preview{}, "", "preview render is missing the selected slot")
	}

	preview, err := s.previewsRepository.GetByJobID(ctx, job.ID)
	if err != nil {
		return s.failRenderJob(ctx, job, models.Preview{}, *job.SelectedSlotID, "preview record is missing")
	}
	slot, err := s.slotsRepository.GetByID(ctx, job.ID, *job.SelectedSlotID)
	if err != nil {
		return err
	}
	campaign, err := s.campaignRepository.GetByID(ctx, job.CampaignID)
	if err != nil {
		return err
	}

	artifactManifest := cloneMetadata(preview.ArtifactManifest)
	if artifactManifest == nil {
		artifactManifest = models.Metadata{}
	}

	sourceBlobURI, err := s.ensurePreviewArtifactUploaded(ctx, job.ID, artifactManifest, "source_video_blob_uri", campaign.VideoPath, fmt.Sprintf("%s/source%s", job.ID, filepath.Ext(campaign.VideoPath)))
	if err != nil {
		return s.failRenderJob(ctx, job, preview, slot.ID, "source video upload failed")
	}
	clipPath := strings.TrimSpace(stringValueOrDefault(slot.GeneratedClipPath))
	clipBlobURI, err := s.ensurePreviewArtifactUploaded(ctx, job.ID, artifactManifest, "generation_blob_uri", clipPath, fmt.Sprintf("%s/generated%s", job.ID, filepath.Ext(clipPath)))
	if err != nil {
		return s.failRenderJob(ctx, job, preview, slot.ID, "generated clip upload failed")
	}

	audioBlobURI := ""
	if slot.GeneratedAudioPath != nil && strings.TrimSpace(*slot.GeneratedAudioPath) != "" {
		audioBlobURI, err = s.ensurePreviewArtifactUploaded(ctx, job.ID, artifactManifest, "generation_audio_blob_uri", strings.TrimSpace(*slot.GeneratedAudioPath), fmt.Sprintf("%s/generated_audio%s", job.ID, filepath.Ext(strings.TrimSpace(*slot.GeneratedAudioPath))))
		if err != nil {
			return s.failRenderJob(ctx, job, preview, slot.ID, "generated audio upload failed")
		}
	}

	if err := s.previewsRepository.MarkStitching(ctx, job.ID, slot.ID, artifactManifest); err != nil {
		return err
	}

	response, err := s.renderClient.SubmitRender(ctx, RenderRequest{
		JobID:                 job.ID,
		SlotID:                slot.ID,
		SourceVideoBlobURI:    sourceBlobURI,
		GeneratedClipBlobURI:  clipBlobURI,
		GeneratedAudioBlobURI: audioBlobURI,
		AnchorStartFrame:      slot.AnchorStartFrame,
		AnchorEndFrame:        slot.AnchorEndFrame,
		SourceFPS:             metadataFloat(job.Metadata, "source_fps"),
		TargetOutputName:      previewOutputFilename(job.ID),
		AudioStrategy:         "crossfade",
	})
	if err != nil {
		fallbackHandled, fallbackErr := s.tryLocalRenderFallback(ctx, job, preview, slot, campaign, artifactManifest, fmt.Sprintf("render submission failed: %v", err))
		if fallbackHandled || fallbackErr != nil {
			return fallbackErr
		}
		return s.failRenderJob(ctx, job, preview, slot.ID, "render submission failed")
	}

	if response.PayloadRef != "" {
		job.Metadata[internalRenderPayloadRef] = response.PayloadRef
	}
	if response.PreviewBlobURI != "" {
		artifactManifest["preview_blob_uri"] = response.PreviewBlobURI
		if err := s.previewsRepository.MarkStitching(ctx, job.ID, slot.ID, artifactManifest); err != nil {
			return err
		}
	}

	switch strings.ToLower(strings.TrimSpace(response.Status)) {
	case "completed", "succeeded":
		return s.persistCompletedRender(ctx, job, preview, slot, campaign, response, artifactManifest)
	case "failed", "error":
		message := strings.TrimSpace(response.Message)
		if message == "" {
			message = "preview render failed"
		}
		fallbackHandled, fallbackErr := s.tryLocalRenderFallback(ctx, job, preview, slot, campaign, artifactManifest, message)
		if fallbackHandled || fallbackErr != nil {
			return fallbackErr
		}
		return s.failRenderJob(ctx, job, preview, slot.ID, message)
	}

	tx, beginErr := s.database.BeginTx(ctx, nil)
	if beginErr != nil {
		return beginErr
	}
	defer func() { _ = tx.Rollback() }()

	jobRepo := db.NewJobsRepository(tx)
	logRepo := db.NewJobLogsRepository(tx)

	current, err := jobRepo.GetByID(ctx, job.ID)
	if err != nil {
		return err
	}
	current.Metadata = ensureJobMetadata(current.Metadata)
	current.Metadata[internalRenderRequestIDKey] = response.RequestID
	if response.PayloadRef != "" {
		current.Metadata[internalRenderPayloadRef] = response.PayloadRef
	}
	current.Status = constants.JobStatusStitching
	current.CurrentStage = constants.StageRenderPoll
	current.ProgressPercent = 80
	if err := jobRepo.UpdateState(ctx, current); err != nil {
		return err
	}
	if err := logRepo.Insert(ctx, models.JobLog{
		JobID:     current.ID,
		Timestamp: TimestampNow(),
		EventType: "stage_started",
		StageName: constants.StageRenderPoll,
		Message:   "submitted preview render request",
	}); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *JobService) processRenderPoll(ctx context.Context, job models.Job) error {
	job.Metadata = ensureJobMetadata(job.Metadata)
	if job.SelectedSlotID == nil || *job.SelectedSlotID == "" {
		return s.failRenderJob(ctx, job, models.Preview{}, "", "preview render is missing the selected slot")
	}

	preview, err := s.previewsRepository.GetByJobID(ctx, job.ID)
	if err != nil {
		return s.failRenderJob(ctx, job, models.Preview{}, *job.SelectedSlotID, "preview record is missing")
	}
	slot, err := s.slotsRepository.GetByID(ctx, job.ID, *job.SelectedSlotID)
	if err != nil {
		return err
	}
	campaign, err := s.campaignRepository.GetByID(ctx, job.CampaignID)
	if err != nil {
		return err
	}
	artifactManifest := cloneMetadata(preview.ArtifactManifest)
	if artifactManifest == nil {
		artifactManifest = models.Metadata{}
	}

	requestID := metadataString(job.Metadata, internalRenderRequestIDKey)
	if requestID == "" {
		return s.failRenderJob(ctx, job, preview, slot.ID, "render request is missing")
	}

	response, err := s.renderClient.PollRender(ctx, RenderPollRequest{
		JobID:     job.ID,
		SlotID:    slot.ID,
		RequestID: requestID,
	})
	if err != nil {
		fallbackHandled, fallbackErr := s.tryLocalRenderFallback(ctx, job, preview, slot, campaign, artifactManifest, fmt.Sprintf("render polling failed: %v", err))
		if fallbackHandled || fallbackErr != nil {
			return fallbackErr
		}
		return s.failRenderJob(ctx, job, preview, slot.ID, "render polling failed")
	}
	if response.PreviewBlobURI != "" {
		artifactManifest["preview_blob_uri"] = response.PreviewBlobURI
	}

	switch strings.ToLower(strings.TrimSpace(response.Status)) {
	case "", "pending", "running", "processing", "submitted":
		if response.PreviewBlobURI != "" {
			return s.previewsRepository.MarkStitching(ctx, job.ID, slot.ID, artifactManifest)
		}
		return nil
	case "failed", "error":
		message := strings.TrimSpace(response.Message)
		if message == "" {
			message = "preview render failed"
		}
		fallbackHandled, fallbackErr := s.tryLocalRenderFallback(ctx, job, preview, slot, campaign, artifactManifest, message)
		if fallbackHandled || fallbackErr != nil {
			return fallbackErr
		}
		return s.failRenderJob(ctx, job, preview, slot.ID, message)
	case "completed", "succeeded":
		return s.persistCompletedRender(ctx, job, preview, slot, campaign, response, artifactManifest)
	default:
		fallbackHandled, fallbackErr := s.tryLocalRenderFallback(ctx, job, preview, slot, campaign, artifactManifest, fmt.Sprintf("render returned unknown status %q", response.Status))
		if fallbackHandled || fallbackErr != nil {
			return fallbackErr
		}
		return s.failRenderJob(ctx, job, preview, slot.ID, fmt.Sprintf("render returned unknown status %q", response.Status))
	}
}

func (s *JobService) persistCompletedRender(ctx context.Context, job models.Job, preview models.Preview, slot models.Slot, campaign models.Campaign, response RenderResponse, artifactManifest models.Metadata) error {
	previewBlobURI := metadataString(artifactManifest, "preview_blob_uri")
	if strings.TrimSpace(response.PreviewBlobURI) != "" {
		previewBlobURI = strings.TrimSpace(response.PreviewBlobURI)
		artifactManifest["preview_blob_uri"] = previewBlobURI
	}
	if previewBlobURI == "" {
		return s.failRenderJob(ctx, job, preview, slot.ID, "render completed without preview artifact")
	}

	download, err := s.blobClient.Download(ctx, BlobDownloadRequest{
		JobID:   job.ID,
		BlobURI: previewBlobURI,
	})
	if err != nil {
		return s.failRenderJob(ctx, job, preview, slot.ID, "preview download failed")
	}
	defer download.Body.Close()

	outputPath := filepath.Join(s.previewDir, previewOutputFilename(job.ID))
	if err := s.storage.SaveReader(outputPath, download.Body); err != nil {
		return s.failRenderJob(ctx, job, preview, slot.ID, "failed to persist preview output")
	}

	insertedDuration := response.DurationSeconds - campaign.DurationSeconds
	if insertedDuration <= 0 {
		insertedDuration = float64(campaign.TargetAdDurationSeconds)
	}
	previewDuration := response.DurationSeconds
	if previewDuration <= 0 {
		previewDuration = campaign.DurationSeconds + insertedDuration
	}

	renderMetrics := cloneMetadata(preview.RenderMetrics)
	if renderMetrics == nil {
		renderMetrics = models.Metadata{}
	}
	renderMetrics["source_duration_seconds"] = roundScore(campaign.DurationSeconds)
	renderMetrics["preview_duration_seconds"] = roundScore(previewDuration)
	renderMetrics["inserted_duration_seconds"] = roundScore(insertedDuration)
	renderMetrics["anchor_start_frame"] = slot.AnchorStartFrame
	renderMetrics["anchor_end_frame"] = slot.AnchorEndFrame

	completedAt := TimestampNow()
	tx, beginErr := s.database.BeginTx(ctx, nil)
	if beginErr != nil {
		return beginErr
	}
	defer func() { _ = tx.Rollback() }()

	jobRepo := db.NewJobsRepository(tx)
	previewRepo := db.NewPreviewsRepository(tx)
	logRepo := db.NewJobLogsRepository(tx)

	current, err := jobRepo.GetByID(ctx, job.ID)
	if err != nil {
		return err
	}
	current.Metadata = ensureJobMetadata(current.Metadata)
	delete(current.Metadata, internalRenderRequestIDKey)
	if response.PayloadRef != "" {
		current.Metadata[internalRenderPayloadRef] = response.PayloadRef
	}
	current.Status = constants.JobStatusCompleted
	current.CurrentStage = constants.StageRenderPoll
	current.ProgressPercent = 100
	current.ErrorCode = nil
	current.ErrorMessage = nil
	current.CompletedAt = &completedAt
	if err := jobRepo.UpdateState(ctx, current); err != nil {
		return err
	}

	if err := previewRepo.MarkCompleted(ctx, models.Preview{
		ID:               preview.ID,
		JobID:            job.ID,
		SlotID:           slot.ID,
		Status:           "completed",
		OutputVideoPath:  outputPath,
		DurationSeconds:  previewDuration,
		RenderRetryCount: preview.RenderRetryCount,
		CreatedAt:        preview.CreatedAt,
		CompletedAt:      &completedAt,
		ArtifactManifest: artifactManifest,
		RenderMetrics:    renderMetrics,
	}); err != nil {
		return err
	}

	if err := logRepo.Insert(ctx, models.JobLog{
		JobID:     current.ID,
		Timestamp: completedAt,
		EventType: "stage_completed",
		StageName: constants.StageRenderPoll,
		Message:   "preview render complete",
	}); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *JobService) failRenderJob(ctx context.Context, job models.Job, preview models.Preview, slotID, message string) error {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	jobRepo := db.NewJobsRepository(tx)
	previewRepo := db.NewPreviewsRepository(tx)
	logRepo := db.NewJobLogsRepository(tx)

	current, err := jobRepo.GetByID(ctx, job.ID)
	if err != nil {
		return err
	}

	if slotID == "" && current.SelectedSlotID != nil {
		slotID = *current.SelectedSlotID
	}

	artifactManifest := cloneMetadata(preview.ArtifactManifest)
	if artifactManifest == nil {
		artifactManifest = models.Metadata{}
	}
	completedAt := TimestampNow()
	if preview.JobID != "" {
		if err := previewRepo.MarkFailed(ctx, current.ID, slotID, constants.ErrorCodePreviewRenderFailed, message, completedAt, true, artifactManifest); err != nil {
			return err
		}
	}

	errorCode := constants.ErrorCodePreviewRenderFailed
	errorMessage := strings.TrimSpace(message)
	if errorMessage == "" {
		errorMessage = "preview render failed"
	}
	current.Status = constants.JobStatusFailed
	current.CurrentStage = constants.StageRender
	current.ProgressPercent = 80
	current.ErrorCode = &errorCode
	current.ErrorMessage = &errorMessage
	current.CompletedAt = &completedAt
	if err := jobRepo.UpdateState(ctx, current); err != nil {
		return err
	}
	if err := logRepo.Insert(ctx, models.JobLog{
		JobID:     current.ID,
		Timestamp: completedAt,
		EventType: "job_failed",
		StageName: constants.StageRender,
		Message:   errorMessage,
	}); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *JobService) ensurePreviewArtifactUploaded(ctx context.Context, jobID string, artifactManifest models.Metadata, key, sourcePath, objectName string) (string, error) {
	if uri := metadataString(artifactManifest, key); uri != "" {
		return uri, nil
	}
	upload, err := s.blobClient.Upload(ctx, BlobUploadRequest{
		JobID:      jobID,
		Path:       sourcePath,
		ObjectName: objectName,
	})
	if err != nil {
		return "", err
	}
	artifactManifest[key] = upload.BlobURI
	return upload.BlobURI, nil
}

func previewOutputFilename(jobID string) string {
	return fmt.Sprintf("%s_preview.mp4", jobID)
}
