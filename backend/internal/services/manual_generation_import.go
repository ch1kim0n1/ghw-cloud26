package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/constants"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/db"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

const GenerationProviderManualImport = "manual_import"

func (s *JobService) ImportManualGeneration(ctx context.Context, jobID, slotID string, startSeconds, endSeconds *float64, generatedClipPath, generatedAudioPath string) (models.Job, models.Slot, error) {
	clipPath := strings.TrimSpace(generatedClipPath)
	if clipPath == "" {
		return models.Job{}, models.Slot{}, InvalidRequest(constants.ErrorCodeInvalidRequest, "generated_clip_path is required", map[string]any{
			"job_id": jobID,
		})
	}
	if err := validateManualImportFile(clipPath, "generated clip"); err != nil {
		return models.Job{}, models.Slot{}, err
	}

	audioPath := strings.TrimSpace(generatedAudioPath)
	if audioPath != "" {
		if err := validateManualImportFile(audioPath, "generated audio"); err != nil {
			return models.Job{}, models.Slot{}, err
		}
	}

	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to start manual generation import transaction", map[string]any{
			"job_id": jobID,
		}, err)
	}
	defer func() { _ = tx.Rollback() }()

	jobRepo := db.NewJobsRepository(tx)
	slotRepo := db.NewSlotsRepository(tx)
	sceneRepo := db.NewScenesRepository(tx)
	logRepo := db.NewJobLogsRepository(tx)

	job, err := jobRepo.GetByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return models.Job{}, models.Slot{}, ResourceNotFound("job not found", map[string]any{
				"job_id": jobID,
			}, err)
		}
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to load job", map[string]any{
			"job_id": jobID,
		}, err)
	}
	if job.Status == constants.JobStatusCompleted || job.Status == constants.JobStatusStitching {
		return models.Job{}, models.Slot{}, Conflict(constants.ErrorCodeInvalidRequest, "manual generation import is not allowed for a completed or stitching job", map[string]any{
			"job_id": jobID,
			"status": job.Status,
		})
	}

	selectedSlot, _ := loadOptionalSelectedSlot(ctx, slotRepo, job)

	targetSlot, jobChanged, err := s.resolveManualImportSlot(ctx, job, selectedSlot, slotRepo, sceneRepo, slotID, startSeconds, endSeconds)
	if err != nil {
		return models.Job{}, models.Slot{}, err
	}

	now := TimestampNow()
	targetSlot.Status = constants.SlotStatusSelected
	targetSlot.GenerationError = nil
	targetSlot.GeneratedClipPath = nil
	targetSlot.GeneratedAudioPath = nil
	targetSlot.UpdatedAt = now
	targetSlot.Metadata = ensureMetadata(targetSlot.Metadata)
	targetSlot.Metadata["manual_generation_import"] = true
	targetSlot.Metadata["manual_generation_source_clip"] = clipPath
	if audioPath != "" {
		targetSlot.Metadata["manual_generation_source_audio"] = audioPath
	}
	if err := slotRepo.Upsert(ctx, targetSlot); err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to persist import slot state", map[string]any{
			"job_id":  jobID,
			"slot_id": targetSlot.ID,
		}, err)
	}

	job.Metadata = ensureJobMetadata(job.Metadata)
	delete(job.Metadata, internalGenerationRequestIDKey)
	delete(job.Metadata, internalGenerationRequestSnapshotKey)
	job.SelectedSlotID = &targetSlot.ID
	job.Status = constants.JobStatusAnalyzing
	job.CurrentStage = constants.StageLineReview
	job.ProgressPercent = 40
	job.ErrorCode = nil
	job.ErrorMessage = nil
	job.CompletedAt = nil
	job.Metadata["generation_provider_attempted"] = GenerationProviderManualImport
	job.Metadata["generation_provider_used"] = GenerationProviderManualImport
	job.Metadata["generation_fallback_used"] = false
	job.Metadata["generation_fallback_reason"] = ""
	if err := jobRepo.UpdateState(ctx, job); err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to update job for manual generation import", map[string]any{
			"job_id":  jobID,
			"slot_id": targetSlot.ID,
		}, err)
	}

	logMessage := "manual generated clip imported"
	if jobChanged {
		logMessage = "manual slot selected and generated clip imported"
	}
	if err := logRepo.Insert(ctx, models.JobLog{
		JobID:     job.ID,
		Timestamp: now,
		EventType: "manual_generation_imported",
		StageName: constants.StageGenerationPoll,
		Message:   logMessage,
	}); err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to write manual import log", map[string]any{
			"job_id":  jobID,
			"slot_id": targetSlot.ID,
		}, err)
	}

	if err := tx.Commit(); err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to commit manual generation import", map[string]any{
			"job_id":  jobID,
			"slot_id": targetSlot.ID,
		}, err)
	}

	importedClipPath, err := s.copyManualImportArtifact(job.ID, targetSlot.ID, clipPath, "generated")
	if err != nil {
		return models.Job{}, models.Slot{}, NewAppError(http.StatusInternalServerError, constants.ErrorCodeStorageError, "failed to persist imported generated clip", map[string]any{
			"job_id":  jobID,
			"slot_id": targetSlot.ID,
			"path":    clipPath,
		}, err)
	}

	importedAudioPath := ""
	if audioPath != "" {
		importedAudioPath, err = s.copyManualImportArtifact(job.ID, targetSlot.ID, audioPath, "generated_audio")
		if err != nil {
			return models.Job{}, models.Slot{}, NewAppError(http.StatusInternalServerError, constants.ErrorCodeStorageError, "failed to persist imported generated audio", map[string]any{
				"job_id":  jobID,
				"slot_id": targetSlot.ID,
				"path":    audioPath,
			}, err)
		}
	}

	if err := s.persistCompletedGeneration(ctx, job, targetSlot, GenerationResponse{
		RequestID:          "manual-import",
		Status:             "completed",
		GeneratedClipPath:  importedClipPath,
		GeneratedAudioPath: importedAudioPath,
		PayloadRef:         "manual-import:" + targetSlot.ID,
		Metadata: models.Metadata{
			"generation_provider_attempted": GenerationProviderManualImport,
			"generation_provider_used":      GenerationProviderManualImport,
			"generation_fallback_used":      false,
			"generation_fallback_reason":    "",
			"manual_generation_import":      true,
			"manual_generation_source_clip": clipPath,
		},
	}); err != nil {
		return models.Job{}, models.Slot{}, err
	}

	refreshedJob, err := s.Get(ctx, jobID)
	if err != nil {
		return models.Job{}, models.Slot{}, err
	}
	refreshedSlot, err := s.GetSlot(ctx, jobID, targetSlot.ID)
	if err != nil {
		return models.Job{}, models.Slot{}, err
	}
	return refreshedJob, refreshedSlot, nil
}

func (s *JobService) resolveManualImportSlot(ctx context.Context, job models.Job, selectedSlot models.Slot, slotRepo *db.SlotsRepository, sceneRepo *db.ScenesRepository, slotID string, startSeconds, endSeconds *float64) (models.Slot, bool, error) {
	if strings.TrimSpace(slotID) != "" {
		slot, err := slotRepo.GetByID(ctx, job.ID, strings.TrimSpace(slotID))
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				return models.Slot{}, false, ResourceNotFound("slot not found", map[string]any{
					"job_id":  job.ID,
					"slot_id": slotID,
				}, err)
			}
			return models.Slot{}, false, DatabaseFailure("failed to load slot", map[string]any{
				"job_id":  job.ID,
				"slot_id": slotID,
			}, err)
		}
		return slot, job.SelectedSlotID == nil || *job.SelectedSlotID != slot.ID, nil
	}

	if startSeconds != nil && endSeconds != nil {
		slot, err := s.buildManualImportSlot(ctx, job, selectedSlot, sceneRepo, *startSeconds, *endSeconds)
		if err != nil {
			return models.Slot{}, false, err
		}
		return slot, true, nil
	}

	if job.SelectedSlotID != nil && *job.SelectedSlotID != "" {
		slot, err := slotRepo.GetByID(ctx, job.ID, *job.SelectedSlotID)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				return models.Slot{}, false, ResourceNotFound("selected slot not found", map[string]any{
					"job_id":  job.ID,
					"slot_id": *job.SelectedSlotID,
				}, err)
			}
			return models.Slot{}, false, DatabaseFailure("failed to load selected slot", map[string]any{
				"job_id":  job.ID,
				"slot_id": *job.SelectedSlotID,
			}, err)
		}
		return slot, false, nil
	}

	return models.Slot{}, false, InvalidRequest(constants.ErrorCodeInvalidRequest, "manual generation import requires either slot_id or start_seconds/end_seconds", map[string]any{
		"job_id": job.ID,
	})
}

func (s *JobService) buildManualImportSlot(ctx context.Context, job models.Job, selectedSlot models.Slot, sceneRepo *db.ScenesRepository, startSeconds, endSeconds float64) (models.Slot, error) {
	if startSeconds < 0 || endSeconds <= 0 || endSeconds <= startSeconds {
		return models.Slot{}, InvalidRequest(constants.ErrorCodeInvalidRequest, "manual generation import requires start_seconds < end_seconds within the source video", map[string]any{
			"job_id":        job.ID,
			"start_seconds": startSeconds,
			"end_seconds":   endSeconds,
		})
	}

	job.Metadata = ensureJobMetadata(job.Metadata)
	durationSeconds := metadataFloat(job.Metadata, "duration_seconds")
	if durationSeconds > 0 && endSeconds > durationSeconds {
		return models.Slot{}, InvalidRequest(constants.ErrorCodeInvalidRequest, "manual generation import must fall within the source video duration", map[string]any{
			"job_id":           job.ID,
			"duration_seconds": durationSeconds,
			"start_seconds":    startSeconds,
			"end_seconds":      endSeconds,
		})
	}

	scenes, err := sceneRepo.ListByJobID(ctx, job.ID)
	if err != nil {
		return models.Slot{}, DatabaseFailure("failed to load scenes", map[string]any{
			"job_id": job.ID,
		}, err)
	}
	scene, ok := findSceneForManualSelection(scenes, startSeconds, endSeconds)
	if !ok && len(scenes) == 0 {
		startFrame, endFrame := sourceSceneFrames(job, startSeconds, endSeconds)
		scene = buildSyntheticManualImportScene(job, startFrame, endFrame)
		if err := sceneRepo.Insert(ctx, scene); err != nil {
			return models.Slot{}, DatabaseFailure("failed to persist synthetic manual import scene", map[string]any{
				"job_id":        job.ID,
				"start_seconds": startSeconds,
				"end_seconds":   endSeconds,
			}, err)
		}
		ok = true
	}
	if !ok {
		return models.Slot{}, InvalidRequest(constants.ErrorCodeInvalidRequest, "manual generation import must stay inside one analyzed scene", map[string]any{
			"job_id":        job.ID,
			"start_seconds": startSeconds,
			"end_seconds":   endSeconds,
		})
	}

	sourceFPS := metadataFloat(job.Metadata, "source_fps")
	if sourceFPS <= 0 {
		sourceFPS = 24
	}
	anchorStartFrame := int(startSeconds * sourceFPS)
	anchorEndFrame := int(endSeconds * sourceFPS)
	if anchorStartFrame < scene.StartFrame {
		anchorStartFrame = scene.StartFrame
	}
	if anchorEndFrame > scene.EndFrame {
		anchorEndFrame = scene.EndFrame
	}
	if anchorEndFrame <= anchorStartFrame {
		return models.Slot{}, InvalidRequest(constants.ErrorCodeInvalidRequest, "manual generation import produced invalid anchor frames for the chosen scene", map[string]any{
			"job_id":        job.ID,
			"start_seconds": startSeconds,
			"end_seconds":   endSeconds,
			"scene_id":      scene.ID,
		})
	}

	quietWindowSeconds := endSeconds - startSeconds
	motionScore := clamp01(floatValue(scene.MotionScore, 0.2))
	stabilityScore := clamp01(floatValue(scene.StabilityScore, clamp01(1-motionScore)))
	dialogueScore := clamp01(floatValue(scene.DialogueActivityScore, 0.3))
	actionIntensity := clamp01(floatValue(scene.ActionIntensityScore, motionScore))
	quietWindowScore := clamp01(quietWindowSeconds / 3)
	narrativeFitScore := roundScore(clamp01((1-dialogueScore)*0.50 + quietWindowScore*0.30 + (1-actionIntensity)*0.20))
	anchorContinuityScore := roundScore(clamp01(stabilityScore*0.40 + (1-motionScore)*0.35 + (1-clamp01(floatValue(scene.AbruptCutRisk, 0.1)))*0.25))
	slotScore := roundScore(stabilityScore*0.35 + quietWindowScore*0.25 + float64(narrativeFitScore)*0.20 + float64(anchorContinuityScore)*0.15)

	slotID := fmt.Sprintf("slot_%s_manual_import_%d_%d", job.ID, int(startSeconds*1000), int(endSeconds*1000))
	now := TimestampNow()
	slot := models.Slot{
		ID:                    slotID,
		JobID:                 job.ID,
		Rank:                  0,
		SceneID:               scene.ID,
		AnchorStartFrame:      anchorStartFrame,
		AnchorEndFrame:        anchorEndFrame,
		SourceFPS:             sourceFPS,
		QuietWindowSeconds:    quietWindowSeconds,
		Score:                 slotScore,
		Reasoning:             "manual selection by operator for imported generation",
		Status:                constants.SlotStatusSelected,
		SuggestedProductLine:  selectedSlot.SuggestedProductLine,
		FinalProductLine:      selectedSlot.FinalProductLine,
		ProductLineMode:       selectedSlot.ProductLineMode,
		NarrativeFitScore:     floatPtr(narrativeFitScore),
		AnchorContinuityScore: floatPtr(anchorContinuityScore),
		Metadata: models.Metadata{
			"manual":                   true,
			"manual_generation_import": true,
			"manual_start_seconds":     startSeconds,
			"manual_end_seconds":       endSeconds,
			"provider_scene":           scene.ID,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	return slot, nil
}

func sourceSceneFrames(job models.Job, startSeconds, endSeconds float64) (int, int) {
	metadata := ensureMetadata(job.Metadata)
	sourceFPS := metadataFloat(metadata, "source_fps")
	if sourceFPS <= 0 {
		sourceFPS = 24
	}
	durationSeconds := metadataFloat(metadata, "duration_seconds")
	startFrame := int(startSeconds * sourceFPS)
	endFrame := int(endSeconds * sourceFPS)
	if startFrame < 0 {
		startFrame = 0
	}
	if durationSeconds > 0 {
		maxFrame := int(durationSeconds * sourceFPS)
		if endFrame > maxFrame {
			endFrame = maxFrame
		}
	}
	if endFrame <= startFrame {
		endFrame = startFrame + 1
	}
	return startFrame, endFrame
}

func buildSyntheticManualImportScene(job models.Job, startFrame, endFrame int) models.Scene {
	metadata := ensureMetadata(job.Metadata)
	sourceFPS := metadataFloat(metadata, "source_fps")
	if sourceFPS <= 0 {
		sourceFPS = 24
	}
	durationSeconds := metadataFloat(metadata, "duration_seconds")
	if durationSeconds <= 0 {
		durationSeconds = float64(endFrame-startFrame) / sourceFPS
	}
	maxFrame := int(durationSeconds * sourceFPS)
	if maxFrame <= 0 {
		maxFrame = endFrame
	}
	return models.Scene{
		ID:                    fmt.Sprintf("scene_%s_manual_import_synthetic", job.ID),
		JobID:                 job.ID,
		SceneNumber:           0,
		StartFrame:            0,
		EndFrame:              maxFrame,
		StartSeconds:          0,
		EndSeconds:            durationSeconds,
		CreatedAt:             TimestampNow(),
		Metadata:              models.Metadata{"synthetic_manual_import_scene": true},
		MotionScore:           floatPtr(0.2),
		StabilityScore:        floatPtr(0.8),
		DialogueActivityScore: floatPtr(0.3),
		ActionIntensityScore:  floatPtr(0.2),
		AbruptCutRisk:         floatPtr(0.1),
	}
}

func (s *JobService) copyManualImportArtifact(jobID, slotID, srcPath, filenamePrefix string) (string, error) {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", srcPath, err)
	}
	defer srcFile.Close()

	ext := filepath.Ext(srcPath)
	if ext == "" {
		ext = ".bin"
	}
	dstPath := filepath.Join(filepath.Dir(s.previewDir), "artifacts", jobID, "manual-import", slotID, filenamePrefix+ext)
	if err := s.storage.SaveReader(dstPath, srcFile); err != nil {
		return "", err
	}
	return dstPath, nil
}

func validateManualImportFile(path, label string) error {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ResourceNotFound(label+" file not found", map[string]any{
				"path": path,
			}, err)
		}
		return NewAppError(http.StatusInternalServerError, constants.ErrorCodeStorageError, "failed to inspect "+label+" file", map[string]any{
			"path": path,
		}, err)
	}
	if info.IsDir() {
		return InvalidRequest(constants.ErrorCodeInvalidRequest, label+" path must be a file", map[string]any{
			"path": path,
		})
	}
	return nil
}

func loadOptionalSelectedSlot(ctx context.Context, slotRepo *db.SlotsRepository, job models.Job) (models.Slot, bool) {
	if job.SelectedSlotID == nil || *job.SelectedSlotID == "" {
		return models.Slot{}, false
	}
	slot, err := slotRepo.GetByID(ctx, job.ID, *job.SelectedSlotID)
	if err != nil {
		return models.Slot{}, false
	}
	return slot, true
}

func ensureMetadata(metadata models.Metadata) models.Metadata {
	if metadata == nil {
		return models.Metadata{}
	}
	return metadata
}
