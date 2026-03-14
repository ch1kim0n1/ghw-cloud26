package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/constants"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/db"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

type suggestedProductLineEnvelope struct {
	SuggestedProductLine string `json:"suggested_product_line"`
}

func (s *JobService) SelectSlot(ctx context.Context, jobID, slotID string) (models.Job, models.Slot, error) {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to start slot selection transaction", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}
	defer func() { _ = tx.Rollback() }()

	jobRepo := db.NewJobsRepository(tx)
	campaignRepo := db.NewCampaignsRepository(tx)
	productRepo := db.NewProductsRepository(tx)
	sceneRepo := db.NewScenesRepository(tx)
	slotRepo := db.NewSlotsRepository(tx)
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
	if job.CurrentStage != constants.StageSlotSelection || job.Status != constants.JobStatusAnalyzing {
		return models.Job{}, models.Slot{}, Conflict(constants.ErrorCodeInvalidRequest, "job is not in slot selection", map[string]any{
			"job_id":        jobID,
			"status":        job.Status,
			"current_stage": job.CurrentStage,
		})
	}

	slot, err := slotRepo.GetByID(ctx, jobID, slotID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return models.Job{}, models.Slot{}, ResourceNotFound("slot not found", map[string]any{
				"job_id":  jobID,
				"slot_id": slotID,
			}, err)
		}
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to load slot", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}
	if slot.Status != constants.SlotStatusProposed {
		return models.Job{}, models.Slot{}, Conflict(constants.ErrorCodeInvalidRequest, "slot cannot be selected from its current state", map[string]any{
			"job_id":      jobID,
			"slot_id":     slotID,
			"slot_status": slot.Status,
		})
	}

	campaign, err := campaignRepo.GetByID(ctx, job.CampaignID)
	if err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to load campaign", map[string]any{
			"job_id": jobID,
		}, err)
	}
	product, err := productRepo.GetByID(ctx, campaign.ProductID)
	if err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to load product", map[string]any{
			"job_id":     jobID,
			"product_id": campaign.ProductID,
		}, err)
	}
	scenes, err := sceneRepo.ListByJobID(ctx, job.ID)
	if err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to load scenes", map[string]any{
			"job_id": jobID,
		}, err)
	}
	scene, ok := findSceneByID(scenes, slot.SceneID)
	if !ok {
		return models.Job{}, models.Slot{}, ResourceNotFound("scene not found for slot", map[string]any{
			"job_id":   jobID,
			"slot_id":  slotID,
			"scene_id": slot.SceneID,
		}, db.ErrNotFound)
	}

	suggestedLine, requestID, err := s.generateSuggestedProductLine(ctx, job.ID, campaign, product, scene, slot, metadataString(job.Metadata, "content_language"))
	if err != nil {
		return models.Job{}, models.Slot{}, NewAppError(http.StatusInternalServerError, constants.ErrorCodeGenerationFailed, "failed to prepare suggested product line", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}

	now := TimestampNow()
	if err := slotRepo.UpdateSelected(ctx, jobID, slotID, suggestedLine, now); err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to update selected slot", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}

	job.Metadata = ensureJobMetadata(job.Metadata)
	if requestID != "" {
		job.Metadata[internalProductLineRequestIDKey] = requestID
	}
	job.Status = constants.JobStatusAnalyzing
	job.CurrentStage = constants.StageLineReview
	job.ProgressPercent = 40
	job.ErrorCode = nil
	job.ErrorMessage = nil
	job.CompletedAt = nil
	job.SelectedSlotID = &slotID
	if err := jobRepo.UpdateState(ctx, job); err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to update job for selected slot", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}
	if err := logRepo.Insert(ctx, models.JobLog{
		JobID:     job.ID,
		Timestamp: now,
		EventType: "slot_selected",
		StageName: constants.StageLineReview,
		Message:   "slot selected and product line prepared",
	}); err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to write slot selection log", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}

	if err := tx.Commit(); err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to commit slot selection", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}

	refreshedSlot, err := s.GetSlot(ctx, jobID, slotID)
	if err != nil {
		return models.Job{}, models.Slot{}, err
	}
	return sanitizeJob(job), refreshedSlot, nil
}

func (s *JobService) SelectManualSlot(ctx context.Context, jobID string, startSeconds, endSeconds float64) (models.Job, models.Slot, error) {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to start manual slot selection transaction", map[string]any{
			"job_id": jobID,
		}, err)
	}
	defer func() { _ = tx.Rollback() }()

	jobRepo := db.NewJobsRepository(tx)
	campaignRepo := db.NewCampaignsRepository(tx)
	productRepo := db.NewProductsRepository(tx)
	sceneRepo := db.NewScenesRepository(tx)
	slotRepo := db.NewSlotsRepository(tx)
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
	if job.CurrentStage != constants.StageSlotSelection || job.Status != constants.JobStatusAnalyzing {
		return models.Job{}, models.Slot{}, Conflict(constants.ErrorCodeInvalidRequest, "job is not in slot selection", map[string]any{
			"job_id":        jobID,
			"status":        job.Status,
			"current_stage": job.CurrentStage,
		})
	}

	if startSeconds < 0 || endSeconds <= 0 || endSeconds <= startSeconds {
		return models.Job{}, models.Slot{}, InvalidRequest(constants.ErrorCodeInvalidRequest, "manual slot selection requires start_seconds < end_seconds within the source video", map[string]any{
			"job_id":        jobID,
			"start_seconds": startSeconds,
			"end_seconds":   endSeconds,
		})
	}

	job.Metadata = ensureJobMetadata(job.Metadata)
	durationSeconds := metadataFloat(job.Metadata, "duration_seconds")
	if durationSeconds > 0 && endSeconds > durationSeconds {
		return models.Job{}, models.Slot{}, InvalidRequest(constants.ErrorCodeInvalidRequest, "manual slot selection must fall within the source video duration", map[string]any{
			"job_id":           jobID,
			"duration_seconds": durationSeconds,
			"start_seconds":    startSeconds,
			"end_seconds":      endSeconds,
		})
	}

	campaign, err := campaignRepo.GetByID(ctx, job.CampaignID)
	if err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to load campaign", map[string]any{
			"job_id": jobID,
		}, err)
	}
	product, err := productRepo.GetByID(ctx, campaign.ProductID)
	if err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to load product", map[string]any{
			"job_id":     jobID,
			"product_id": campaign.ProductID,
		}, err)
	}
	scenes, err := sceneRepo.ListByJobID(ctx, job.ID)
	if err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to load scenes", map[string]any{
			"job_id": jobID,
		}, err)
	}
	scene, ok := findSceneForManualSelection(scenes, startSeconds, endSeconds)
	if !ok {
		return models.Job{}, models.Slot{}, InvalidRequest(constants.ErrorCodeInvalidRequest, "manual slot selection must stay inside one analyzed scene", map[string]any{
			"job_id":        jobID,
			"start_seconds": startSeconds,
			"end_seconds":   endSeconds,
		})
	}

	sourceFPS := metadataFloat(job.Metadata, "source_fps")
	if sourceFPS <= 0 {
		sourceFPS = 24
	}
	anchorStartFrame := int(math.Round(startSeconds * sourceFPS))
	anchorEndFrame := int(math.Round(endSeconds * sourceFPS))
	if anchorStartFrame < scene.StartFrame {
		anchorStartFrame = scene.StartFrame
	}
	if anchorEndFrame > scene.EndFrame {
		anchorEndFrame = scene.EndFrame
	}
	if anchorEndFrame <= anchorStartFrame {
		anchorEndFrame = anchorStartFrame + 1
		if anchorEndFrame > scene.EndFrame {
			return models.Job{}, models.Slot{}, InvalidRequest(constants.ErrorCodeInvalidRequest, "manual slot selection produced invalid anchor frames for the chosen scene", map[string]any{
				"job_id":        jobID,
				"start_seconds": startSeconds,
				"end_seconds":   endSeconds,
				"scene_id":      scene.ID,
			})
		}
	}

	quietWindowSeconds := endSeconds - startSeconds
	contextRelevanceScore := roundScore(scoreContextRelevance(product, scene))
	motionScore := clamp01(floatValue(scene.MotionScore, 0.2))
	stabilityScore := clamp01(floatValue(scene.StabilityScore, clamp01(1-motionScore)))
	dialogueScore := clamp01(floatValue(scene.DialogueActivityScore, 0.3))
	actionIntensity := clamp01(floatValue(scene.ActionIntensityScore, motionScore))
	quietWindowScore := clamp01(quietWindowSeconds / 3)
	narrativeFitScore := roundScore(clamp01((1-dialogueScore)*0.50 + quietWindowScore*0.30 + (1-actionIntensity)*0.20))
	anchorContinuityScore := roundScore(clamp01(stabilityScore*0.40 + (1-motionScore)*0.35 + (1-clamp01(floatValue(scene.AbruptCutRisk, 0.1)))*0.25))
	slotScore := roundScore(stabilityScore*0.35 + quietWindowScore*0.25 + float64(narrativeFitScore)*0.20 + float64(anchorContinuityScore)*0.15 + float64(contextRelevanceScore)*0.05)

	slotID := fmt.Sprintf("slot_%s_manual_%d_%d", jobID, int(math.Round(startSeconds*1000)), int(math.Round(endSeconds*1000)))
	suggestedLine, requestID, err := s.generateSuggestedProductLine(ctx, job.ID, campaign, product, scene, models.Slot{
		ID:                 slotID,
		JobID:              jobID,
		SceneID:            scene.ID,
		AnchorStartFrame:   anchorStartFrame,
		AnchorEndFrame:     anchorEndFrame,
		QuietWindowSeconds: quietWindowSeconds,
		Score:              slotScore,
		Reasoning:          "manual selection by operator",
		Status:             constants.SlotStatusSelected,
	}, metadataString(job.Metadata, "content_language"))
	if err != nil {
		return models.Job{}, models.Slot{}, NewAppError(http.StatusInternalServerError, constants.ErrorCodeGenerationFailed, "failed to prepare suggested product line", map[string]any{
			"job_id": jobID,
		}, err)
	}

	now := TimestampNow()
	manualSlot := models.Slot{
		ID:                    slotID,
		JobID:                 jobID,
		Rank:                  0,
		SceneID:               scene.ID,
		AnchorStartFrame:      anchorStartFrame,
		AnchorEndFrame:        anchorEndFrame,
		SourceFPS:             sourceFPS,
		QuietWindowSeconds:    quietWindowSeconds,
		Score:                 slotScore,
		Reasoning:             "manual selection by operator",
		Status:                constants.SlotStatusSelected,
		SuggestedProductLine:  ptrIfNotEmpty(suggestedLine),
		ContextRelevanceScore: floatPtr(contextRelevanceScore),
		NarrativeFitScore:     floatPtr(narrativeFitScore),
		AnchorContinuityScore: floatPtr(anchorContinuityScore),
		Metadata: models.Metadata{
			"manual":               true,
			"manual_start_seconds": roundScore(startSeconds),
			"manual_end_seconds":   roundScore(endSeconds),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := slotRepo.Upsert(ctx, manualSlot); err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to persist manual slot", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}

	if requestID != "" {
		job.Metadata[internalProductLineRequestIDKey] = requestID
	}
	job.Status = constants.JobStatusAnalyzing
	job.CurrentStage = constants.StageLineReview
	job.ProgressPercent = 40
	job.ErrorCode = nil
	job.ErrorMessage = nil
	job.CompletedAt = nil
	job.SelectedSlotID = &slotID
	if err := jobRepo.UpdateState(ctx, job); err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to update job for manual slot selection", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}
	if err := logRepo.Insert(ctx, models.JobLog{
		JobID:     job.ID,
		Timestamp: now,
		EventType: "slot_selected",
		StageName: constants.StageLineReview,
		Message:   "manual slot selected and product line prepared",
	}); err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to write manual slot selection log", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}

	if err := tx.Commit(); err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to commit manual slot selection", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}

	refreshedSlot, err := s.GetSlot(ctx, jobID, slotID)
	if err != nil {
		return models.Job{}, models.Slot{}, err
	}
	return sanitizeJob(job), refreshedSlot, nil
}

func (s *JobService) StartGeneration(ctx context.Context, jobID, slotID, productLineMode, customProductLine string) (models.Job, models.Slot, error) {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to start generation transaction", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}
	defer func() { _ = tx.Rollback() }()

	jobRepo := db.NewJobsRepository(tx)
	slotRepo := db.NewSlotsRepository(tx)
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
	if job.CurrentStage != constants.StageLineReview || job.Status != constants.JobStatusAnalyzing {
		return models.Job{}, models.Slot{}, Conflict(constants.ErrorCodeInvalidRequest, "job is not in product line review", map[string]any{
			"job_id":        jobID,
			"status":        job.Status,
			"current_stage": job.CurrentStage,
		})
	}
	if job.SelectedSlotID == nil || *job.SelectedSlotID != slotID {
		return models.Job{}, models.Slot{}, Conflict(constants.ErrorCodeInvalidRequest, "generation can only start for the selected slot", map[string]any{
			"job_id":           jobID,
			"slot_id":          slotID,
			"selected_slot_id": job.SelectedSlotID,
		})
	}

	slot, err := slotRepo.GetByID(ctx, jobID, slotID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return models.Job{}, models.Slot{}, ResourceNotFound("slot not found", map[string]any{
				"job_id":  jobID,
				"slot_id": slotID,
			}, err)
		}
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to load slot", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}
	if slot.Status != constants.SlotStatusSelected {
		return models.Job{}, models.Slot{}, Conflict(constants.ErrorCodeInvalidRequest, "generation can only start for a selected slot", map[string]any{
			"job_id":      jobID,
			"slot_id":     slotID,
			"slot_status": slot.Status,
		})
	}

	finalProductLine, normalizedMode, validationErr := resolveFinalProductLine(slot, productLineMode, customProductLine)
	if validationErr != nil {
		return models.Job{}, models.Slot{}, validationErr
	}

	now := TimestampNow()
	if err := slotRepo.UpdateGenerationStarted(ctx, jobID, slotID, normalizedMode, finalProductLine, now); err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to update slot generation state", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}

	job.Metadata = ensureJobMetadata(job.Metadata)
	job.Status = constants.JobStatusGenerating
	job.CurrentStage = constants.StageGenerationSubmit
	job.ProgressPercent = 40
	job.ErrorCode = nil
	job.ErrorMessage = nil
	job.CompletedAt = nil
	if err := jobRepo.UpdateState(ctx, job); err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to update job generation state", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}
	if err := logRepo.Insert(ctx, models.JobLog{
		JobID:     job.ID,
		Timestamp: now,
		EventType: "stage_started",
		StageName: constants.StageGenerationSubmit,
		Message:   "cafai generation started",
	}); err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to write generation start log", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}

	if err := tx.Commit(); err != nil {
		return models.Job{}, models.Slot{}, DatabaseFailure("failed to commit generation start", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}

	refreshedSlot, err := s.GetSlot(ctx, jobID, slotID)
	if err != nil {
		return models.Job{}, models.Slot{}, err
	}
	return sanitizeJob(job), refreshedSlot, nil
}

func (s *JobService) processGenerationSubmission(ctx context.Context, job models.Job) error {
	job.Metadata = ensureJobMetadata(job.Metadata)
	if requestID := metadataString(job.Metadata, internalGenerationRequestIDKey); requestID != "" {
		return nil
	}
	if job.SelectedSlotID == nil || *job.SelectedSlotID == "" {
		return s.failGenerationJob(ctx, job, "", constants.StageGenerationSubmit, "generation request is missing the selected slot")
	}

	campaign, err := s.campaignRepository.GetByID(ctx, job.CampaignID)
	if err != nil {
		return err
	}
	product, err := s.productsRepository.GetByID(ctx, campaign.ProductID)
	if err != nil {
		return err
	}
	slot, err := s.slotsRepository.GetByID(ctx, job.ID, *job.SelectedSlotID)
	if err != nil {
		return err
	}
	scenes, err := s.scenesRepository.ListByJobID(ctx, job.ID)
	if err != nil {
		return err
	}
	scene, ok := findSceneByID(scenes, slot.SceneID)
	if !ok {
		return s.failGenerationJob(ctx, job, slot.ID, constants.StageGenerationSubmit, "selected slot scene is missing")
	}

	contentLanguage := metadataString(job.Metadata, "content_language")
	videoHash, productHash, fingerprintErr := s.cacheFingerprints(campaign, product)
	if fingerprintErr != nil {
		return s.failGenerationJob(ctx, job, slot.ID, constants.StageGenerationSubmit, summarizeProviderFailure("generation cache fingerprint failed", fingerprintErr))
	}
	cacheKeys := generationOutputCacheKeys(s.mlClient, s.cache, videoHash, productHash, scene, slot, contentLanguage, strings.TrimSpace(stringValueOrDefault(slot.FinalProductLine)))
	for _, generationOutputCacheKey := range cacheKeys {
		if cachedOutput, ok, cacheErr := s.cache.LoadGenerationOutput(generationOutputCacheKey); cacheErr != nil {
			return s.failGenerationJob(ctx, job, slot.ID, constants.StageGenerationSubmit, summarizeProviderFailure("generation output cache read failed", cacheErr))
		} else if ok && cachedOutput.GeneratedClipPath != "" {
			if _, statErr := os.Stat(cachedOutput.GeneratedClipPath); statErr == nil {
				return s.persistCompletedGeneration(ctx, job, slot, GenerationResponse{
					RequestID:          "cache",
					Status:             "completed",
					GeneratedClipPath:  cachedOutput.GeneratedClipPath,
					GeneratedAudioPath: cachedOutput.GeneratedAudioPath,
					PayloadRef:         "cache:" + generationOutputCacheKey,
					Metadata:           cloneMetadata(cachedOutput.Metadata),
				})
			}
		}
	}

	anchorFrames, err := s.frameExtractor.Extract(ctx, AnchorFrameRequest{
		JobID:            job.ID,
		SlotID:           slot.ID,
		VideoPath:        campaign.VideoPath,
		AnchorStartFrame: slot.AnchorStartFrame,
		AnchorEndFrame:   slot.AnchorEndFrame,
		SourceFPS:        metadataFloat(job.Metadata, "source_fps"),
	})
	if err != nil {
		return s.failGenerationJob(ctx, job, slot.ID, constants.StageGenerationSubmit, "anchor frame extraction failed")
	}

	generationBrief, generationBriefRequestID, err := s.generateGenerationBrief(ctx, job.ID, campaign, product, scene, slot, anchorFrames, contentLanguage)
	if err != nil {
		return s.failGenerationJob(ctx, job, slot.ID, constants.StageGenerationSubmit, summarizeProviderFailure("generation brief preparation failed", err))
	}

	generationRequest := GenerationRequest{
		JobID:                   job.ID,
		SlotID:                  slot.ID,
		CampaignID:              campaign.ID,
		ProductID:               product.ID,
		SceneID:                 scene.ID,
		SourceVideoPath:         campaign.VideoPath,
		AnchorStartImagePath:    anchorFrames.AnchorStartImagePath,
		AnchorEndImagePath:      anchorFrames.AnchorEndImagePath,
		AnchorStartFrame:        slot.AnchorStartFrame,
		AnchorEndFrame:          slot.AnchorEndFrame,
		SourceFPS:               metadataFloat(job.Metadata, "source_fps"),
		TargetDurationSeconds:   campaign.TargetAdDurationSeconds,
		ProductName:             product.Name,
		ProductDescription:      product.Description,
		ProductCategory:         product.Category,
		ProductContextKeywords:  product.ContextKeywords,
		ProductImagePath:        product.ImagePath,
		ProductSourceURL:        product.SourceURL,
		SceneNarrativeSummary:   scene.NarrativeSummary,
		ProductLineMode:         strings.TrimSpace(stringValueOrDefault(slot.ProductLineMode)),
		SuggestedProductLine:    strings.TrimSpace(stringValueOrDefault(slot.SuggestedProductLine)),
		FinalProductLine:        strings.TrimSpace(stringValueOrDefault(slot.FinalProductLine)),
		GenerationBrief:         generationBrief,
		ContentLanguage:         contentLanguage,
		SelectedSlotReasoning:   slot.Reasoning,
		SelectedSlotQuietWindow: slot.QuietWindowSeconds,
	}

	response, err := s.mlClient.SubmitGeneration(ctx, generationRequest)
	if err != nil {
		return s.failGenerationJob(ctx, job, slot.ID, constants.StageGenerationSubmit, summarizeProviderFailure("generation submission failed", err))
	}

	switch strings.ToLower(strings.TrimSpace(response.Status)) {
	case "completed", "succeeded":
		return s.persistCompletedGeneration(ctx, job, slot, response)
	case "failed", "error":
		message := strings.TrimSpace(response.Message)
		if message == "" {
			message = "generation failed"
		}
		return s.failGenerationJob(ctx, job, slot.ID, constants.StageGenerationSubmit, message)
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
	setGenerationProviderMetadata(current.Metadata, response.Metadata)
	if generationBriefRequestID != "" {
		current.Metadata[internalGenerationBriefRequestIDKey] = generationBriefRequestID
	}
	if requestSnapshot, snapshotErr := encodeGenerationRequestSnapshot(generationRequest); snapshotErr == nil {
		current.Metadata[internalGenerationRequestSnapshotKey] = requestSnapshot
	}
	current.Metadata[internalGenerationRequestIDKey] = response.RequestID
	if response.PayloadRef != "" {
		current.Metadata[internalGenerationPayloadRef] = response.PayloadRef
	}
	current.Status = constants.JobStatusGenerating
	current.CurrentStage = constants.StageGenerationPoll
	current.ProgressPercent = 40
	if err := jobRepo.UpdateState(ctx, current); err != nil {
		return err
	}
	if err := logRepo.Insert(ctx, models.JobLog{
		JobID:     current.ID,
		Timestamp: TimestampNow(),
		EventType: "stage_started",
		StageName: constants.StageGenerationPoll,
		Message:   "submitted cafai generation request",
	}); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *JobService) processGenerationPoll(ctx context.Context, job models.Job) error {
	job.Metadata = ensureJobMetadata(job.Metadata)
	if job.SelectedSlotID == nil || *job.SelectedSlotID == "" {
		return nil
	}

	slot, err := s.slotsRepository.GetByID(ctx, job.ID, *job.SelectedSlotID)
	if err != nil {
		return err
	}

	requestID := metadataString(job.Metadata, internalGenerationRequestIDKey)
	if requestID == "" {
		if slot.Status == constants.SlotStatusGenerated {
			return nil
		}
		return s.failGenerationJob(ctx, job, slot.ID, constants.StageGenerationPoll, "generation request is missing")
	}

	response, err := s.mlClient.PollGeneration(ctx, GenerationPollRequest{
		JobID:     job.ID,
		SlotID:    slot.ID,
		RequestID: requestID,
	})
	if err != nil {
		fallbackHandled, fallbackErr := s.tryGenerationFallback(ctx, job, slot, fmt.Sprintf("generation polling failed: %v", err))
		if fallbackHandled || fallbackErr != nil {
			return fallbackErr
		}
		return s.failGenerationJob(ctx, job, slot.ID, constants.StageGenerationPoll, "generation polling failed")
	}

	switch strings.ToLower(strings.TrimSpace(response.Status)) {
	case "", "pending", "running", "processing", "submitted":
		return nil
	case "failed", "error":
		message := strings.TrimSpace(response.Message)
		if message == "" {
			message = "generation failed"
		}
		fallbackHandled, fallbackErr := s.tryGenerationFallback(ctx, job, slot, message)
		if fallbackHandled || fallbackErr != nil {
			return fallbackErr
		}
		return s.failGenerationJob(ctx, job, slot.ID, constants.StageGenerationPoll, message)
	case "completed", "succeeded":
		return s.persistCompletedGeneration(ctx, job, slot, response)
	default:
		return s.failGenerationJob(ctx, job, slot.ID, constants.StageGenerationPoll, fmt.Sprintf("generation returned unknown status %q", response.Status))
	}
}

func (s *JobService) persistCompletedGeneration(ctx context.Context, job models.Job, slot models.Slot, response GenerationResponse) error {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	jobRepo := db.NewJobsRepository(tx)
	slotRepo := db.NewSlotsRepository(tx)
	logRepo := db.NewJobLogsRepository(tx)

	current, err := jobRepo.GetByID(ctx, job.ID)
	if err != nil {
		return err
	}
	currentSlot, err := slotRepo.GetByID(ctx, job.ID, slot.ID)
	if err != nil {
		return err
	}

	current.Metadata = ensureJobMetadata(current.Metadata)
	delete(current.Metadata, internalGenerationRequestIDKey)
	delete(current.Metadata, internalGenerationRequestSnapshotKey)
	if response.PayloadRef != "" {
		current.Metadata[internalGenerationPayloadRef] = response.PayloadRef
	}
	setGenerationProviderMetadata(current.Metadata, response.Metadata)

	slotMetadata := sanitizeGenerationMetadata(currentSlot.Metadata, response.Metadata)
	clipPath := ptrIfNotEmpty(response.GeneratedClipPath)
	audioPath := ptrIfNotEmpty(response.GeneratedAudioPath)
	if err := slotRepo.UpdateGenerationSucceeded(ctx, current.ID, currentSlot.ID, clipPath, audioPath, slotMetadata, TimestampNow()); err != nil {
		return err
	}

	current.Status = constants.JobStatusGenerating
	current.CurrentStage = constants.StageGenerationPoll
	current.ProgressPercent = 80
	current.ErrorCode = nil
	current.ErrorMessage = nil
	current.CompletedAt = nil
	if err := jobRepo.UpdateState(ctx, current); err != nil {
		return err
	}
	if err := logRepo.Insert(ctx, models.JobLog{
		JobID:     current.ID,
		Timestamp: TimestampNow(),
		EventType: "stage_completed",
		StageName: constants.StageGenerationPoll,
		Message:   "cafai generation complete",
	}); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	_ = s.saveGenerationOutputCache(ctx, current, currentSlot, clipPath, audioPath, slotMetadata)
	return nil
}

func (s *JobService) failGenerationJob(ctx context.Context, job models.Job, slotID, stage, message string) error {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	jobRepo := db.NewJobsRepository(tx)
	slotRepo := db.NewSlotsRepository(tx)
	logRepo := db.NewJobLogsRepository(tx)

	current, err := jobRepo.GetByID(ctx, job.ID)
	if err != nil {
		return err
	}
	current.Metadata = ensureJobMetadata(current.Metadata)
	delete(current.Metadata, internalGenerationRequestIDKey)
	delete(current.Metadata, internalGenerationRequestSnapshotKey)

	if slotID == "" && current.SelectedSlotID != nil {
		slotID = *current.SelectedSlotID
	}
	if slotID != "" {
		if err := slotRepo.UpdateGenerationFailed(ctx, current.ID, slotID, message, TimestampNow()); err != nil {
			return err
		}
	}

	errorCode := constants.ErrorCodeGenerationFailed
	errorMessage := strings.TrimSpace(message)
	if errorMessage == "" {
		errorMessage = "generation failed"
	}
	completedAt := TimestampNow()
	current.Status = constants.JobStatusFailed
	current.CurrentStage = stage
	current.ProgressPercent = 40
	current.ErrorCode = &errorCode
	current.ErrorMessage = &errorMessage
	current.CompletedAt = &completedAt
	if err := jobRepo.UpdateState(ctx, current); err != nil {
		return err
	}
	if err := logRepo.Insert(ctx, models.JobLog{
		JobID:     current.ID,
		Timestamp: TimestampNow(),
		EventType: "job_failed",
		StageName: stage,
		Message:   errorMessage,
	}); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *JobService) tryGenerationFallback(ctx context.Context, job models.Job, slot models.Slot, reason string) (bool, error) {
	if metadataString(job.Metadata, "generation_provider_used") != GenerationProviderHiggsfield {
		return false, nil
	}
	if metadataBool(job.Metadata, "generation_fallback_used") {
		return false, nil
	}

	fallbackClient, ok := s.mlClient.(FallbackGenerationSubmitter)
	if !ok {
		return false, nil
	}

	requestSnapshot := metadataString(job.Metadata, internalGenerationRequestSnapshotKey)
	if requestSnapshot == "" {
		return false, nil
	}
	generationRequest, err := decodeGenerationRequestSnapshot(requestSnapshot)
	if err != nil {
		return false, s.failGenerationJob(ctx, job, slot.ID, constants.StageGenerationPoll, summarizeProviderFailure("generation fallback request decode failed", err))
	}

	response, err := fallbackClient.SubmitGenerationFallback(ctx, generationRequest, reason)
	if err != nil {
		return false, s.failGenerationJob(ctx, job, slot.ID, constants.StageGenerationPoll, summarizeProviderFailure("generation fallback submission failed", err))
	}

	switch strings.ToLower(strings.TrimSpace(response.Status)) {
	case "completed", "succeeded":
		return true, s.persistCompletedGeneration(ctx, job, slot, response)
	case "failed", "error":
		message := strings.TrimSpace(response.Message)
		if message == "" {
			message = "generation fallback failed"
		}
		return false, s.failGenerationJob(ctx, job, slot.ID, constants.StageGenerationPoll, message)
	}

	tx, beginErr := s.database.BeginTx(ctx, nil)
	if beginErr != nil {
		return false, beginErr
	}
	defer func() { _ = tx.Rollback() }()

	jobRepo := db.NewJobsRepository(tx)
	logRepo := db.NewJobLogsRepository(tx)

	current, err := jobRepo.GetByID(ctx, job.ID)
	if err != nil {
		return false, err
	}
	current.Metadata = ensureJobMetadata(current.Metadata)
	setGenerationProviderMetadata(current.Metadata, response.Metadata)
	current.Metadata[internalGenerationRequestIDKey] = response.RequestID
	if response.PayloadRef != "" {
		current.Metadata[internalGenerationPayloadRef] = response.PayloadRef
	}
	current.Status = constants.JobStatusGenerating
	current.CurrentStage = constants.StageGenerationPoll
	current.ProgressPercent = 40
	current.ErrorCode = nil
	current.ErrorMessage = nil
	current.CompletedAt = nil
	if err := jobRepo.UpdateState(ctx, current); err != nil {
		return false, err
	}
	if err := logRepo.Insert(ctx, models.JobLog{
		JobID:     current.ID,
		Timestamp: TimestampNow(),
		EventType: "stage_started",
		StageName: constants.StageGenerationPoll,
		Message:   "submitted generation fallback request",
	}); err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

func (s *JobService) generateSuggestedProductLine(ctx context.Context, jobID string, campaign models.Campaign, product models.Product, scene models.Scene, slot models.Slot, contentLanguage string) (string, string, error) {
	videoHash, productHash, err := s.cacheFingerprints(campaign, product)
	if err != nil {
		return "", "", err
	}
	cacheKey := s.cache.PromptKey("suggested-line", videoHash, productHash, cacheSceneKey(scene), cacheSlotKey(slot), contentLanguage)
	if cachedLine, ok, cacheErr := s.cache.LoadSuggestedLine(cacheKey); cacheErr != nil {
		return "", "", cacheErr
	} else if ok {
		return cachedLine, "cache", nil
	}

	prompt, err := buildSuggestedProductLinePrompt(product, scene, slot, contentLanguage)
	if err != nil {
		return "", "", err
	}

	response, err := s.openAIClient.Complete(ctx, OpenAIRequest{
		JobID:        jobID,
		Purpose:      "phase_3_product_line",
		SystemPrompt: suggestedProductLineSystemPrompt(contentLanguage),
		Prompt:       prompt,
		Temperature:  0.4,
	})
	if err != nil {
		return "", "", err
	}

	line, err := parseSuggestedProductLine(response.Content)
	if err != nil {
		return "", response.RequestID, err
	}
	if err := s.cache.SaveSuggestedLine(cacheKey, line); err != nil {
		return "", response.RequestID, err
	}
	return line, response.RequestID, nil
}

func (s *JobService) generateGenerationBrief(ctx context.Context, jobID string, campaign models.Campaign, product models.Product, scene models.Scene, slot models.Slot, anchors AnchorFrameArtifacts, contentLanguage string) (string, string, error) {
	videoHash, productHash, err := s.cacheFingerprints(campaign, product)
	if err != nil {
		return "", "", err
	}
	cacheKey := s.cache.PromptKey(
		"generation-brief",
		videoHash,
		productHash,
		cacheSceneKey(scene),
		cacheSlotKey(slot),
		contentLanguage,
		strings.TrimSpace(stringValueOrDefault(slot.ProductLineMode)),
		strings.TrimSpace(stringValueOrDefault(slot.SuggestedProductLine)),
		strings.TrimSpace(stringValueOrDefault(slot.FinalProductLine)),
	)
	if cachedBrief, ok, cacheErr := s.cache.LoadGenerationBrief(cacheKey); cacheErr != nil {
		return "", "", cacheErr
	} else if ok {
		return cachedBrief, "cache", nil
	}

	prompt, err := buildGenerationBriefPrompt(product, scene, slot, anchors, contentLanguage)
	if err != nil {
		return "", "", err
	}

	response, err := s.openAIClient.Complete(ctx, OpenAIRequest{
		JobID:        jobID,
		Purpose:      "phase_3_generation_brief",
		SystemPrompt: generationBriefSystemPrompt(contentLanguage),
		Prompt:       prompt,
		Temperature:  0.2,
	})
	if err != nil {
		return "", "", err
	}

	brief, err := parseGenerationBrief(response.Content)
	if err != nil {
		return "", response.RequestID, err
	}
	if err := s.cache.SaveGenerationBrief(cacheKey, brief); err != nil {
		return "", response.RequestID, err
	}
	return brief, response.RequestID, nil
}

func buildSuggestedProductLinePrompt(product models.Product, scene models.Scene, slot models.Slot, contentLanguage string) (string, error) {
	payload := map[string]any{
		"content_language": normalizeContentLanguage(contentLanguage),
		"product": map[string]any{
			"name":             product.Name,
			"description":      product.Description,
			"category":         product.Category,
			"context_keywords": product.ContextKeywords,
			"source_url":       product.SourceURL,
		},
		"scene": map[string]any{
			"id":                scene.ID,
			"narrative_summary": scene.NarrativeSummary,
			"context_keywords":  scene.ContextKeywords,
			"start_seconds":     scene.StartSeconds,
			"end_seconds":       scene.EndSeconds,
		},
		"slot": map[string]any{
			"id":                   slot.ID,
			"reasoning":            slot.Reasoning,
			"quiet_window_seconds": slot.QuietWindowSeconds,
		},
		"requirements": map[string]any{
			"style":              "brief natural spoken line",
			"length_words_max":   18,
			"length_words_min":   4,
			"tone":               "plausible in-scene dialogue, not ad copy",
			"allow_product_name": true,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func buildGenerationBriefPrompt(product models.Product, scene models.Scene, slot models.Slot, anchors AnchorFrameArtifacts, contentLanguage string) (string, error) {
	payload := map[string]any{
		"content_language": normalizeContentLanguage(contentLanguage),
		"product": map[string]any{
			"name":             product.Name,
			"description":      product.Description,
			"category":         product.Category,
			"context_keywords": product.ContextKeywords,
			"source_url":       product.SourceURL,
			"image_path":       product.ImagePath,
		},
		"scene": map[string]any{
			"id":                scene.ID,
			"narrative_summary": scene.NarrativeSummary,
			"context_keywords":  scene.ContextKeywords,
			"start_seconds":     scene.StartSeconds,
			"end_seconds":       scene.EndSeconds,
		},
		"slot": map[string]any{
			"id":                     slot.ID,
			"anchor_start_frame":     slot.AnchorStartFrame,
			"anchor_end_frame":       slot.AnchorEndFrame,
			"quiet_window_seconds":   slot.QuietWindowSeconds,
			"reasoning":              slot.Reasoning,
			"suggested_product_line": stringValueOrDefault(slot.SuggestedProductLine),
			"final_product_line":     stringValueOrDefault(slot.FinalProductLine),
			"product_line_mode":      stringValueOrDefault(slot.ProductLineMode),
		},
		"anchors": map[string]any{
			"anchor_start_image_path": anchors.AnchorStartImagePath,
			"anchor_end_image_path":   anchors.AnchorEndImagePath,
		},
		"requirements": map[string]any{
			"duration_seconds_default": 6,
			"duration_seconds_max":     8,
			"goal":                     "describe a short product interaction bridge clip that starts from the start anchor and resolves into the end anchor",
			"format":                   "plain text brief",
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func suggestedProductLineSystemPrompt(contentLanguage string) string {
	return strings.Join([]string{
		"You are preparing one short suggested product mention for a CAFAI insertion moment.",
		"Return strict JSON with the top-level key suggested_product_line.",
		"Write one natural line that could plausibly be spoken in the scene.",
		fmt.Sprintf("Write the spoken line in %s.", normalizeContentLanguage(contentLanguage)),
		"Keep the wording concise and conversational, never salesy.",
		"Do not wrap the JSON in markdown code fences.",
	}, " ")
}

func generationBriefSystemPrompt(contentLanguage string) string {
	return strings.Join([]string{
		"You are preparing a CAFAI generation brief for Azure Machine Learning.",
		"Return strict JSON with the top-level key generation_brief.",
		"Describe a concise 5-8 second bridge clip that starts from the start anchor image, introduces the product naturally, and resolves into the end anchor image.",
		"Reference the selected product line mode and spoken line when present.",
		fmt.Sprintf("Write the generation brief in %s so it matches the source content language.", normalizeContentLanguage(contentLanguage)),
		"Do not wrap the JSON in markdown code fences.",
	}, " ")
}

func parseSuggestedProductLine(content string) (string, error) {
	return parseLLMTextResponse(content, "suggested_product_line", "product_line", "line")
}

func parseGenerationBrief(content string) (string, error) {
	return parseLLMTextResponse(content, "generation_brief", "brief", "content")
}

func setGenerationProviderMetadata(target models.Metadata, provider models.Metadata) {
	if target == nil {
		return
	}
	for _, key := range []string{
		"generation_provider_attempted",
		"generation_provider_used",
		"generation_fallback_used",
		"generation_fallback_reason",
		"higgsfield_model_id",
	} {
		if value, ok := provider[key]; ok {
			target[key] = value
		}
	}
}

func resolveFinalProductLine(slot models.Slot, productLineMode, customProductLine string) (*string, string, error) {
	mode := strings.ToLower(strings.TrimSpace(productLineMode))
	switch mode {
	case "auto":
		if slot.SuggestedProductLine == nil || strings.TrimSpace(*slot.SuggestedProductLine) == "" {
			return nil, "", Conflict(constants.ErrorCodeInvalidRequest, "auto mode requires a suggested product line", map[string]any{
				"slot_id": slot.ID,
			})
		}
		finalLine := strings.TrimSpace(*slot.SuggestedProductLine)
		return &finalLine, mode, nil
	case "operator":
		finalLine := strings.TrimSpace(customProductLine)
		if finalLine == "" {
			return nil, "", InvalidRequest(constants.ErrorCodeInvalidRequest, "custom product line is required for operator mode", map[string]any{
				"slot_id": slot.ID,
			})
		}
		return &finalLine, mode, nil
	case "disabled":
		return nil, mode, nil
	default:
		return nil, "", InvalidRequest(constants.ErrorCodeInvalidRequest, "product_line_mode must be one of auto, operator, or disabled", map[string]any{
			"slot_id": slot.ID,
		})
	}
}

func findSceneByID(scenes []models.Scene, sceneID string) (models.Scene, bool) {
	for _, scene := range scenes {
		if scene.ID == sceneID {
			return scene, true
		}
	}
	return models.Scene{}, false
}

func findSceneForManualSelection(scenes []models.Scene, startSeconds, endSeconds float64) (models.Scene, bool) {
	for _, scene := range scenes {
		if startSeconds >= scene.StartSeconds && endSeconds <= scene.EndSeconds {
			return scene, true
		}
	}
	return models.Scene{}, false
}

func sanitizeGenerationMetadata(existing models.Metadata, provider models.Metadata) models.Metadata {
	metadata := cloneMetadata(existing)
	if metadata == nil {
		metadata = models.Metadata{}
	}
	for key, value := range provider {
		switch key {
		case "request_id", "provider_request_id", "payload_ref", "provider_payload_ref":
			continue
		default:
			metadata[key] = value
		}
	}
	return metadata
}

func ptrIfNotEmpty(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func stringValueOrDefault(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func parseLLMTextResponse(content string, preferredKeys ...string) (string, error) {
	trimmed := strings.TrimSpace(content)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)
	if trimmed == "" {
		return "", fmt.Errorf("llm response was empty")
	}

	var decoded any
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		if err := json.Unmarshal([]byte(trimmed), &decoded); err == nil {
			if text := extractTextFromLLMResponse(decoded, preferredKeys...); text != "" {
				return text, nil
			}
		}
	}

	fallback := strings.TrimSpace(strings.Trim(trimmed, "\""))
	if fallback == "" {
		return "", fmt.Errorf("llm response was empty after normalization")
	}
	return fallback, nil
}

func extractTextFromLLMResponse(value any, preferredKeys ...string) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(strings.Trim(typed, "\""))
	case map[string]any:
		for _, key := range preferredKeys {
			if extracted := extractTextFromLLMResponse(typed[key], preferredKeys...); extracted != "" {
				return extracted
			}
		}
		for _, key := range []string{"message", "text", "output"} {
			if extracted := extractTextFromLLMResponse(typed[key], preferredKeys...); extracted != "" {
				return extracted
			}
		}
	case []any:
		for _, item := range typed {
			if extracted := extractTextFromLLMResponse(item, preferredKeys...); extracted != "" {
				return extracted
			}
		}
	}
	return ""
}

func (s *JobService) saveGenerationOutputCache(ctx context.Context, job models.Job, slot models.Slot, clipPath, audioPath *string, metadata models.Metadata) error {
	if !s.cache.Enabled() || clipPath == nil || *clipPath == "" {
		return nil
	}
	campaign, err := s.campaignRepository.GetByID(ctx, job.CampaignID)
	if err != nil {
		return err
	}
	product, err := s.productsRepository.GetByID(ctx, campaign.ProductID)
	if err != nil {
		return err
	}
	scenes, err := s.scenesRepository.ListByJobID(ctx, job.ID)
	if err != nil {
		return err
	}
	scene, ok := findSceneByID(scenes, slot.SceneID)
	if !ok {
		return nil
	}
	videoHash, productHash, err := s.cacheFingerprints(campaign, product)
	if err != nil {
		return err
	}
	cacheProvider := strings.TrimSpace(metadataString(metadata, "generation_provider_used"))
	if cacheProvider == "" {
		providers := generationCacheProviderNames(s.mlClient)
		cacheProvider = providers[0]
	}
	cacheKey := generationOutputCacheKey(
		s.cache,
		cacheProvider,
		videoHash,
		productHash,
		scene,
		slot,
		metadataString(job.Metadata, "content_language"),
		strings.TrimSpace(stringValueOrDefault(slot.FinalProductLine)),
	)
	return s.cache.SaveGenerationOutput(cacheKey, cachedGenerationOutput{
		GeneratedClipPath:  stringValueOrDefault(clipPath),
		GeneratedAudioPath: stringValueOrDefault(audioPath),
		Metadata:           cloneMetadata(metadata),
	})
}

func generationOutputCacheKeys(client MLClient, cache *ProviderCache, videoHash, productHash string, scene models.Scene, slot models.Slot, contentLanguage, finalProductLine string) []string {
	keys := make([]string, 0, len(generationCacheProviderNames(client)))
	for _, provider := range generationCacheProviderNames(client) {
		keys = append(keys, generationOutputCacheKey(cache, provider, videoHash, productHash, scene, slot, contentLanguage, finalProductLine))
	}
	return keys
}

func generationOutputCacheKey(cache *ProviderCache, provider, videoHash, productHash string, scene models.Scene, slot models.Slot, contentLanguage, finalProductLine string) string {
	return cache.PromptKey(
		"generation-output",
		provider,
		videoHash,
		productHash,
		cacheSceneKey(scene),
		cacheSlotKey(slot),
		contentLanguage,
		finalProductLine,
	)
}

func generationCacheProviderNames(client MLClient) []string {
	switch typed := client.(type) {
	case *PriorityFallbackMLClient:
		return []string{typed.primaryName, typed.fallbackName}
	case *HiggsfieldClient:
		return []string{GenerationProviderHiggsfield}
	case *AzureMLClient:
		return []string{GenerationProviderAzureML}
	case *VultrGenerationClient:
		return []string{GenerationProviderVultr}
	default:
		return []string{"default"}
	}
}

func encodeGenerationRequestSnapshot(req GenerationRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(body), nil
}

func decodeGenerationRequestSnapshot(value string) (GenerationRequest, error) {
	body, err := base64.StdEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		return GenerationRequest{}, err
	}
	var req GenerationRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return GenerationRequest{}, err
	}
	return req, nil
}

func metadataBool(metadata models.Metadata, key string) bool {
	switch value := metadata[key].(type) {
	case bool:
		return value
	case string:
		return strings.EqualFold(strings.TrimSpace(value), "true")
	default:
		return false
	}
}
