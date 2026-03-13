package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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

	suggestedLine, requestID, err := s.generateSuggestedProductLine(ctx, job.ID, product, scene, slot)
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

	generationBrief, generationBriefRequestID, err := s.generateGenerationBrief(ctx, job.ID, product, scene, slot, anchorFrames)
	if err != nil {
		return s.failGenerationJob(ctx, job, slot.ID, constants.StageGenerationSubmit, "generation brief preparation failed")
	}

	response, err := s.mlClient.SubmitGeneration(ctx, GenerationRequest{
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
		SelectedSlotReasoning:   slot.Reasoning,
		SelectedSlotQuietWindow: slot.QuietWindowSeconds,
	})
	if err != nil {
		return s.failGenerationJob(ctx, job, slot.ID, constants.StageGenerationSubmit, "generation submission failed")
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
	if generationBriefRequestID != "" {
		current.Metadata[internalGenerationBriefRequestIDKey] = generationBriefRequestID
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

	return tx.Commit()
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
	if response.PayloadRef != "" {
		current.Metadata[internalGenerationPayloadRef] = response.PayloadRef
	}

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

	return tx.Commit()
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

func (s *JobService) generateSuggestedProductLine(ctx context.Context, jobID string, product models.Product, scene models.Scene, slot models.Slot) (string, string, error) {
	prompt, err := buildSuggestedProductLinePrompt(product, scene, slot)
	if err != nil {
		return "", "", err
	}

	response, err := s.openAIClient.Complete(ctx, OpenAIRequest{
		JobID:        jobID,
		Purpose:      "phase_3_product_line",
		SystemPrompt: suggestedProductLineSystemPrompt(),
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
	return line, response.RequestID, nil
}

func (s *JobService) generateGenerationBrief(ctx context.Context, jobID string, product models.Product, scene models.Scene, slot models.Slot, anchors AnchorFrameArtifacts) (string, string, error) {
	prompt, err := buildGenerationBriefPrompt(product, scene, slot, anchors)
	if err != nil {
		return "", "", err
	}

	response, err := s.openAIClient.Complete(ctx, OpenAIRequest{
		JobID:        jobID,
		Purpose:      "phase_3_generation_brief",
		SystemPrompt: generationBriefSystemPrompt(),
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
	return brief, response.RequestID, nil
}

func buildSuggestedProductLinePrompt(product models.Product, scene models.Scene, slot models.Slot) (string, error) {
	payload := map[string]any{
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

func buildGenerationBriefPrompt(product models.Product, scene models.Scene, slot models.Slot, anchors AnchorFrameArtifacts) (string, error) {
	payload := map[string]any{
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

func suggestedProductLineSystemPrompt() string {
	return strings.Join([]string{
		"You are preparing one short suggested product mention for a CAFAI insertion moment.",
		"Return strict JSON with the top-level key suggested_product_line.",
		"Write one natural line that could plausibly be spoken in the scene.",
		"Keep the wording concise and conversational, never salesy.",
		"Do not wrap the JSON in markdown code fences.",
	}, " ")
}

func generationBriefSystemPrompt() string {
	return strings.Join([]string{
		"You are preparing a CAFAI generation brief for Azure Machine Learning.",
		"Return strict JSON with the top-level key generation_brief.",
		"Describe a concise 5-8 second bridge clip that starts from the start anchor image, introduces the product naturally, and resolves into the end anchor image.",
		"Reference the selected product line mode and spoken line when present.",
		"Do not wrap the JSON in markdown code fences.",
	}, " ")
}

func parseSuggestedProductLine(content string) (string, error) {
	trimmed := strings.TrimSpace(content)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)

	var envelope suggestedProductLineEnvelope
	if strings.HasPrefix(trimmed, "{") {
		if err := json.Unmarshal([]byte(trimmed), &envelope); err != nil {
			return "", fmt.Errorf("decode suggested product line response: %w", err)
		}
		trimmed = envelope.SuggestedProductLine
	}

	trimmed = strings.TrimSpace(trimmed)
	trimmed = strings.Trim(trimmed, "\"")
	if trimmed == "" {
		return "", fmt.Errorf("suggested product line response was empty")
	}
	return trimmed, nil
}

func parseGenerationBrief(content string) (string, error) {
	trimmed := strings.TrimSpace(content)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)

	var envelope struct {
		GenerationBrief string `json:"generation_brief"`
	}
	if strings.HasPrefix(trimmed, "{") {
		if err := json.Unmarshal([]byte(trimmed), &envelope); err != nil {
			return "", fmt.Errorf("decode generation brief response: %w", err)
		}
		trimmed = envelope.GenerationBrief
	}

	trimmed = strings.TrimSpace(trimmed)
	trimmed = strings.Trim(trimmed, "\"")
	if trimmed == "" {
		return "", fmt.Errorf("generation brief response was empty")
	}
	return trimmed, nil
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
