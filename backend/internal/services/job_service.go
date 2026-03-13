package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"slices"
	"strings"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/constants"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/db"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

const (
	internalAnalysisRequestIDKey        = "analysis_request_id"
	internalAnalysisPayloadRef          = "analysis_payload_ref"
	internalProductLineRequestIDKey     = "product_line_request_id"
	internalGenerationBriefRequestIDKey = "generation_brief_request_id"
	internalGenerationRequestIDKey      = "generation_request_id"
	internalGenerationPayloadRef        = "generation_payload_ref"
	internalRenderRequestIDKey          = "render_request_id"
	internalRenderPayloadRef            = "render_payload_ref"
)

type JobService struct {
	database           *sql.DB
	jobsRepository     *db.JobsRepository
	campaignRepository *db.CampaignsRepository
	productsRepository *db.ProductsRepository
	jobLogsRepository  *db.JobLogsRepository
	scenesRepository   *db.ScenesRepository
	slotsRepository    *db.SlotsRepository
	previewsRepository *db.PreviewsRepository
	analysisClient     AnalysisClient
	openAIClient       OpenAIClient
	mlClient           MLClient
	frameExtractor     AnchorFrameExtractor
	storage            *LocalStorageService
	blobClient         BlobStorageClient
	renderClient       RenderClient
	previewDir         string
}

func NewJobService(
	database *sql.DB,
	jobsRepository *db.JobsRepository,
	campaignRepository *db.CampaignsRepository,
	productsRepository *db.ProductsRepository,
	jobLogsRepository *db.JobLogsRepository,
	scenesRepository *db.ScenesRepository,
	slotsRepository *db.SlotsRepository,
	previewsRepository *db.PreviewsRepository,
	analysisClient AnalysisClient,
	openAIClient OpenAIClient,
	mlClient MLClient,
	frameExtractor AnchorFrameExtractor,
	storage *LocalStorageService,
	blobClient BlobStorageClient,
	renderClient RenderClient,
	previewDir string,
) *JobService {
	return &JobService{
		database:           database,
		jobsRepository:     jobsRepository,
		campaignRepository: campaignRepository,
		productsRepository: productsRepository,
		jobLogsRepository:  jobLogsRepository,
		scenesRepository:   scenesRepository,
		slotsRepository:    slotsRepository,
		previewsRepository: previewsRepository,
		analysisClient:     analysisClient,
		openAIClient:       openAIClient,
		mlClient:           mlClient,
		frameExtractor:     frameExtractor,
		storage:            storage,
		blobClient:         blobClient,
		renderClient:       renderClient,
		previewDir:         previewDir,
	}
}

func (s *JobService) Get(ctx context.Context, jobID string) (models.Job, error) {
	job, err := s.jobsRepository.GetByID(ctx, jobID)
	if errors.Is(err, db.ErrNotFound) {
		return models.Job{}, ResourceNotFound("job not found", map[string]any{
			"job_id": jobID,
		}, err)
	}
	if err != nil {
		return models.Job{}, DatabaseFailure("failed to load job", map[string]any{
			"job_id": jobID,
		}, err)
	}

	return sanitizeJob(job), nil
}

func (s *JobService) ListLogs(ctx context.Context, jobID string) ([]models.JobLog, error) {
	if _, err := s.jobsRepository.GetByID(ctx, jobID); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, ResourceNotFound("job not found", map[string]any{
				"job_id": jobID,
			}, err)
		}
		return nil, DatabaseFailure("failed to load job", map[string]any{
			"job_id": jobID,
		}, err)
	}

	logs, err := s.jobLogsRepository.ListByJobID(ctx, jobID)
	if err != nil {
		return nil, DatabaseFailure("failed to load job logs", map[string]any{
			"job_id": jobID,
		}, err)
	}
	return logs, nil
}

func (s *JobService) GetPreview(ctx context.Context, jobID string) (models.Preview, error) {
	if _, err := s.jobsRepository.GetByID(ctx, jobID); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return models.Preview{}, ResourceNotFound("job not found", map[string]any{
				"job_id": jobID,
			}, err)
		}
		return models.Preview{}, DatabaseFailure("failed to load job", map[string]any{
			"job_id": jobID,
		}, err)
	}

	preview, err := s.previewsRepository.GetByJobID(ctx, jobID)
	if errors.Is(err, db.ErrNotFound) {
		return models.Preview{}, ResourceNotFound("preview not found", map[string]any{
			"job_id": jobID,
		}, err)
	}
	if err != nil {
		return models.Preview{}, DatabaseFailure("failed to load preview", map[string]any{
			"job_id": jobID,
		}, err)
	}

	return sanitizePreview(preview), nil
}

func (s *JobService) StartPreviewRender(ctx context.Context, jobID, slotID string) (models.Job, models.Preview, error) {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return models.Job{}, models.Preview{}, DatabaseFailure("failed to start preview render transaction", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}
	defer func() { _ = tx.Rollback() }()

	jobRepo := db.NewJobsRepository(tx)
	slotRepo := db.NewSlotsRepository(tx)
	previewRepo := db.NewPreviewsRepository(tx)
	logRepo := db.NewJobLogsRepository(tx)

	job, err := jobRepo.GetByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return models.Job{}, models.Preview{}, ResourceNotFound("job not found", map[string]any{
				"job_id": jobID,
			}, err)
		}
		return models.Job{}, models.Preview{}, DatabaseFailure("failed to load job", map[string]any{
			"job_id": jobID,
		}, err)
	}

	slot, err := slotRepo.GetByID(ctx, jobID, slotID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return models.Job{}, models.Preview{}, ResourceNotFound("slot not found", map[string]any{
				"job_id":  jobID,
				"slot_id": slotID,
			}, err)
		}
		return models.Job{}, models.Preview{}, DatabaseFailure("failed to load slot", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}
	if job.SelectedSlotID == nil || *job.SelectedSlotID != slotID {
		return models.Job{}, models.Preview{}, Conflict(constants.ErrorCodeInvalidRequest, "preview render can only start for the selected slot", map[string]any{
			"job_id":           jobID,
			"slot_id":          slotID,
			"selected_slot_id": job.SelectedSlotID,
		})
	}
	if slot.Status != constants.SlotStatusGenerated {
		return models.Job{}, models.Preview{}, Conflict(constants.ErrorCodeInvalidRequest, "preview render requires a generated slot", map[string]any{
			"job_id":      jobID,
			"slot_id":     slotID,
			"slot_status": slot.Status,
		})
	}
	if slot.GeneratedClipPath == nil || strings.TrimSpace(*slot.GeneratedClipPath) == "" {
		return models.Job{}, models.Preview{}, Conflict(constants.ErrorCodeInvalidRequest, "generated slot is missing clip output", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		})
	}

	now := TimestampNow()
	previewID := NewPrefixedID("preview")
	preview := models.Preview{
		ID:               previewID,
		JobID:            jobID,
		SlotID:           slotID,
		Status:           "pending",
		CreatedAt:        now,
		ArtifactManifest: models.Metadata{},
		RenderMetrics:    models.Metadata{},
	}

	existingPreview, previewErr := previewRepo.GetByJobID(ctx, jobID)
	switch {
	case previewErr == nil:
		preview.ID = existingPreview.ID
		preview.CreatedAt = existingPreview.CreatedAt
		preview.RenderRetryCount = existingPreview.RenderRetryCount
		preview.ArtifactManifest = cloneMetadata(existingPreview.ArtifactManifest)
		if preview.ArtifactManifest == nil {
			preview.ArtifactManifest = models.Metadata{}
		}
		if existingPreview.Status == "completed" {
			return models.Job{}, models.Preview{}, Conflict(constants.ErrorCodeInvalidRequest, "preview has already been completed for this job", map[string]any{
				"job_id": jobID,
			})
		}
		if existingPreview.Status != "failed" && existingPreview.Status != "pending" && existingPreview.Status != "stitching" {
			return models.Job{}, models.Preview{}, Conflict(constants.ErrorCodeInvalidRequest, "preview render cannot start from the current preview state", map[string]any{
				"job_id":         jobID,
				"preview_status": existingPreview.Status,
			})
		}
	case errors.Is(previewErr, db.ErrNotFound):
	default:
		return models.Job{}, models.Preview{}, DatabaseFailure("failed to load preview", map[string]any{
			"job_id": jobID,
		}, previewErr)
	}

	if err := previewRepo.UpsertPending(ctx, preview); err != nil {
		return models.Job{}, models.Preview{}, DatabaseFailure("failed to persist preview state", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}

	job.Metadata = ensureJobMetadata(job.Metadata)
	delete(job.Metadata, internalRenderRequestIDKey)
	delete(job.Metadata, internalRenderPayloadRef)
	job.Status = constants.JobStatusStitching
	job.CurrentStage = constants.StageRenderSubmit
	job.ProgressPercent = 80
	job.ErrorCode = nil
	job.ErrorMessage = nil
	job.CompletedAt = nil
	if err := jobRepo.UpdateState(ctx, job); err != nil {
		return models.Job{}, models.Preview{}, DatabaseFailure("failed to update job preview state", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}
	if err := logRepo.Insert(ctx, models.JobLog{
		JobID:     job.ID,
		Timestamp: now,
		EventType: "stage_started",
		StageName: constants.StageRenderSubmit,
		Message:   "preview render started",
	}); err != nil {
		return models.Job{}, models.Preview{}, DatabaseFailure("failed to write preview start log", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}

	if err := tx.Commit(); err != nil {
		return models.Job{}, models.Preview{}, DatabaseFailure("failed to commit preview start", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}

	refreshedPreview, err := s.GetPreview(ctx, jobID)
	if err != nil {
		return models.Job{}, models.Preview{}, err
	}
	return sanitizeJob(job), refreshedPreview, nil
}

func (s *JobService) StartAnalysis(ctx context.Context, jobID string) (models.Job, error) {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return models.Job{}, DatabaseFailure("failed to start analysis transaction", map[string]any{
			"job_id": jobID,
		}, err)
	}
	defer func() { _ = tx.Rollback() }()

	jobRepo := db.NewJobsRepository(tx)
	logRepo := db.NewJobLogsRepository(tx)

	job, err := jobRepo.GetByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return models.Job{}, ResourceNotFound("job not found", map[string]any{
				"job_id": jobID,
			}, err)
		}
		return models.Job{}, DatabaseFailure("failed to load job", map[string]any{
			"job_id": jobID,
		}, err)
	}
	if job.Status != constants.JobStatusQueued || job.CurrentStage != constants.StageReadyForAnalysis {
		return models.Job{}, Conflict(constants.ErrorCodeInvalidRequest, "job is not ready for analysis", map[string]any{
			"job_id":        jobID,
			"status":        job.Status,
			"current_stage": job.CurrentStage,
		})
	}

	startedAt := TimestampNow()
	job.Status = constants.JobStatusAnalyzing
	job.CurrentStage = constants.StageAnalysisSubmission
	job.ProgressPercent = 0
	job.StartedAt = &startedAt
	job.CompletedAt = nil
	job.ErrorCode = nil
	job.ErrorMessage = nil
	job.Metadata = ensureJobMetadata(job.Metadata)

	if err := jobRepo.UpdateState(ctx, job); err != nil {
		return models.Job{}, DatabaseFailure("failed to update job", map[string]any{
			"job_id": jobID,
		}, err)
	}
	if err := logRepo.Insert(ctx, models.JobLog{
		JobID:     job.ID,
		Timestamp: TimestampNow(),
		EventType: "stage_started",
		StageName: constants.StageAnalysisSubmission,
		Message:   "analysis started",
	}); err != nil {
		return models.Job{}, DatabaseFailure("failed to write job log", map[string]any{
			"job_id": jobID,
		}, err)
	}

	if err := tx.Commit(); err != nil {
		return models.Job{}, DatabaseFailure("failed to commit analysis start", map[string]any{
			"job_id": jobID,
		}, err)
	}

	return sanitizeJob(job), nil
}

func (s *JobService) ListSlots(ctx context.Context, jobID string) ([]models.Slot, error) {
	job, err := s.jobsRepository.GetByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, ResourceNotFound("job not found", map[string]any{
				"job_id": jobID,
			}, err)
		}
		return nil, DatabaseFailure("failed to load job", map[string]any{
			"job_id": jobID,
		}, err)
	}

	slots, err := s.slotsRepository.ListByJobID(ctx, jobID)
	if err != nil {
		return nil, DatabaseFailure("failed to load slots", map[string]any{
			"job_id": jobID,
		}, err)
	}

	sourceFPS := metadataFloat(job.Metadata, "source_fps")
	for index := range slots {
		slots[index].SourceFPS = sourceFPS
	}

	return slots, nil
}

func (s *JobService) GetSlot(ctx context.Context, jobID, slotID string) (models.Slot, error) {
	job, err := s.jobsRepository.GetByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return models.Slot{}, ResourceNotFound("job not found", map[string]any{
				"job_id": jobID,
			}, err)
		}
		return models.Slot{}, DatabaseFailure("failed to load job", map[string]any{
			"job_id": jobID,
		}, err)
	}

	slot, err := s.slotsRepository.GetByID(ctx, jobID, slotID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return models.Slot{}, ResourceNotFound("slot not found", map[string]any{
				"job_id":  jobID,
				"slot_id": slotID,
			}, err)
		}
		return models.Slot{}, DatabaseFailure("failed to load slot", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}

	slot.SourceFPS = metadataFloat(job.Metadata, "source_fps")
	return slot, nil
}

func (s *JobService) RejectSlot(ctx context.Context, jobID, slotID, note string) (models.Slot, error) {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return models.Slot{}, DatabaseFailure("failed to start slot rejection transaction", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}
	defer func() { _ = tx.Rollback() }()

	jobRepo := db.NewJobsRepository(tx)
	slotRepo := db.NewSlotsRepository(tx)

	job, err := jobRepo.GetByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return models.Slot{}, ResourceNotFound("job not found", map[string]any{
				"job_id": jobID,
			}, err)
		}
		return models.Slot{}, DatabaseFailure("failed to load job", map[string]any{
			"job_id": jobID,
		}, err)
	}
	if job.CurrentStage != constants.StageSlotSelection {
		return models.Slot{}, Conflict(constants.ErrorCodeInvalidRequest, "job is not in slot selection", map[string]any{
			"job_id":        jobID,
			"current_stage": job.CurrentStage,
		})
	}

	slot, err := slotRepo.GetByID(ctx, jobID, slotID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return models.Slot{}, ResourceNotFound("slot not found", map[string]any{
				"job_id":  jobID,
				"slot_id": slotID,
			}, err)
		}
		return models.Slot{}, DatabaseFailure("failed to load slot", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}
	if slot.Status != constants.SlotStatusProposed {
		return models.Slot{}, Conflict(constants.ErrorCodeInvalidRequest, "slot cannot be rejected from its current state", map[string]any{
			"job_id":      jobID,
			"slot_id":     slotID,
			"slot_status": slot.Status,
		})
	}

	trimmedNote := strings.TrimSpace(note)
	if err := slotRepo.UpdateRejected(ctx, jobID, slotID, trimmedNote, TimestampNow()); err != nil {
		return models.Slot{}, DatabaseFailure("failed to reject slot", map[string]any{
			"job_id":  jobID,
			"slot_id": slotID,
		}, err)
	}

	slots, err := slotRepo.ListByJobID(ctx, jobID)
	if err != nil {
		return models.Slot{}, DatabaseFailure("failed to refresh slots", map[string]any{
			"job_id": jobID,
		}, err)
	}

	job.Metadata = ensureJobMetadata(job.Metadata)
	job.Metadata["rejected_slot_ids"] = appendUnique(stringSliceMetadata(job.Metadata, "rejected_slot_ids"), slotID)
	job.Metadata["top_slot_ids"] = proposedSlotIDs(slots)
	if err := jobRepo.UpdateState(ctx, job); err != nil {
		return models.Slot{}, DatabaseFailure("failed to update job metadata", map[string]any{
			"job_id": jobID,
		}, err)
	}

	if err := db.NewJobLogsRepository(tx).Insert(ctx, models.JobLog{
		JobID:     jobID,
		Timestamp: TimestampNow(),
		EventType: "slot_rejected",
		StageName: constants.StageSlotSelection,
		Message:   "slot rejected",
	}); err != nil {
		return models.Slot{}, DatabaseFailure("failed to write slot rejection log", map[string]any{
			"job_id": jobID,
		}, err)
	}

	if err := tx.Commit(); err != nil {
		return models.Slot{}, DatabaseFailure("failed to commit slot rejection", map[string]any{
			"job_id": jobID,
		}, err)
	}

	refreshedSlot, err := s.GetSlot(ctx, jobID, slotID)
	if err != nil {
		return models.Slot{}, err
	}
	if trimmedNote != "" {
		if refreshedSlot.Metadata == nil {
			refreshedSlot.Metadata = models.Metadata{}
		}
		refreshedSlot.Metadata["rejection_note"] = trimmedNote
	}
	return refreshedSlot, nil
}

func (s *JobService) RequestRepick(ctx context.Context, jobID string) (models.Job, error) {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return models.Job{}, DatabaseFailure("failed to start re-pick transaction", map[string]any{
			"job_id": jobID,
		}, err)
	}
	defer func() { _ = tx.Rollback() }()

	jobRepo := db.NewJobsRepository(tx)
	slotRepo := db.NewSlotsRepository(tx)
	sceneRepo := db.NewScenesRepository(tx)
	logRepo := db.NewJobLogsRepository(tx)
	campaignRepo := db.NewCampaignsRepository(tx)
	productRepo := db.NewProductsRepository(tx)

	job, err := jobRepo.GetByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return models.Job{}, ResourceNotFound("job not found", map[string]any{
				"job_id": jobID,
			}, err)
		}
		return models.Job{}, DatabaseFailure("failed to load job", map[string]any{
			"job_id": jobID,
		}, err)
	}
	if job.CurrentStage != constants.StageSlotSelection {
		return models.Job{}, Conflict(constants.ErrorCodeInvalidRequest, "job is not in slot selection", map[string]any{
			"job_id":        jobID,
			"current_stage": job.CurrentStage,
		})
	}

	currentSlots, err := slotRepo.ListByJobID(ctx, jobID)
	if err != nil {
		return models.Job{}, DatabaseFailure("failed to load current slots", map[string]any{
			"job_id": jobID,
		}, err)
	}
	if !canRepick(currentSlots, metadataRepickCount(job.Metadata)) {
		return models.Job{}, Conflict(constants.ErrorCodeInvalidRequest, "re-pick requires all current proposed slots to be rejected", map[string]any{
			"job_id": jobID,
		})
	}

	campaign, err := campaignRepo.GetByID(ctx, job.CampaignID)
	if err != nil {
		return models.Job{}, DatabaseFailure("failed to load campaign", map[string]any{
			"job_id": jobID,
		}, err)
	}
	product, err := productRepo.GetByID(ctx, campaign.ProductID)
	if err != nil {
		return models.Job{}, DatabaseFailure("failed to load product", map[string]any{
			"job_id":     jobID,
			"product_id": campaign.ProductID,
		}, err)
	}
	scenes, err := sceneRepo.ListByJobID(ctx, jobID)
	if err != nil {
		return models.Job{}, DatabaseFailure("failed to load scenes", map[string]any{
			"job_id": jobID,
		}, err)
	}

	job.Metadata = ensureJobMetadata(job.Metadata)
	rejectedIDs := stringSliceMetadata(job.Metadata, "rejected_slot_ids")
	for _, slot := range currentSlots {
		if slot.Status == constants.SlotStatusRejected {
			rejectedIDs = appendUnique(rejectedIDs, slot.ID)
		}
	}
	job.Metadata["rejected_slot_ids"] = rejectedIDs
	nextRepickCount := metadataRepickCount(job.Metadata) + 1
	job.Metadata["repick_count"] = nextRepickCount

	slots, rankingRequestID, rankErr := s.rankSlotsWithOpenAI(ctx, job.ID, metadataFloat(job.Metadata, "source_fps"), product, scenes, rejectedIDs)
	if rankErr != nil {
		if failErr := s.failAnalysisJob(ctx, job, constants.StageSlotSelection, "slot re-pick failed"); failErr != nil {
			return models.Job{}, failErr
		}
		return models.Job{}, NewAppError(http.StatusInternalServerError, constants.ErrorCodeAnalysisFailed, "slot re-pick failed", map[string]any{
			"job_id": jobID,
		}, rankErr)
	}
	if len(slots) == 0 {
		job.Metadata["top_slot_ids"] = []string{}
		job.ProgressPercent = 40
		job.CurrentStage = constants.StageSlotSelection
		if nextRepickCount >= 2 {
			errorCode := constants.ErrorCodeNoSuitableSlot
			errorMessage := "no suitable slot found after re-pick attempts"
			completedAt := TimestampNow()
			job.Status = constants.JobStatusFailed
			job.ErrorCode = &errorCode
			job.ErrorMessage = &errorMessage
			job.CompletedAt = &completedAt
			if err := logRepo.Insert(ctx, models.JobLog{
				JobID:     jobID,
				Timestamp: TimestampNow(),
				EventType: "job_failed",
				StageName: constants.StageSlotSelection,
				Message:   "no suitable slot found after re-pick attempts",
			}); err != nil {
				return models.Job{}, DatabaseFailure("failed to write failure log", map[string]any{
					"job_id": jobID,
				}, err)
			}
		} else {
			job.Status = constants.JobStatusAnalyzing
			job.ErrorCode = nil
			job.ErrorMessage = nil
			job.CompletedAt = nil
			if err := logRepo.Insert(ctx, models.JobLog{
				JobID:     jobID,
				Timestamp: TimestampNow(),
				EventType: "repick_requested",
				StageName: constants.StageSlotSelection,
				Message:   "re-pick requested but no additional valid slots were found",
			}); err != nil {
				return models.Job{}, DatabaseFailure("failed to write re-pick log", map[string]any{
					"job_id": jobID,
				}, err)
			}
		}
		if err := jobRepo.UpdateState(ctx, job); err != nil {
			return models.Job{}, DatabaseFailure("failed to update job after re-pick", map[string]any{
				"job_id": jobID,
			}, err)
		}
		if err := tx.Commit(); err != nil {
			return models.Job{}, DatabaseFailure("failed to commit re-pick", map[string]any{
				"job_id": jobID,
			}, err)
		}
		return sanitizeJob(job), nil
	}

	job.Status = constants.JobStatusAnalyzing
	job.CurrentStage = constants.StageSlotSelection
	job.ProgressPercent = 40
	job.ErrorCode = nil
	job.ErrorMessage = nil
	job.CompletedAt = nil
	job.Metadata["top_slot_ids"] = slotIDs(slots)
	if rankingRequestID != "" {
		job.Metadata[internalSlotRankingRequestIDKey] = rankingRequestID
	}

	if err := slotRepo.ReplaceForJob(ctx, jobID, slots); err != nil {
		return models.Job{}, DatabaseFailure("failed to replace slot set", map[string]any{
			"job_id": jobID,
		}, err)
	}
	if err := jobRepo.UpdateState(ctx, job); err != nil {
		return models.Job{}, DatabaseFailure("failed to update job after re-pick", map[string]any{
			"job_id": jobID,
		}, err)
	}
	if err := logRepo.Insert(ctx, models.JobLog{
		JobID:     jobID,
		Timestamp: TimestampNow(),
		EventType: "repick_requested",
		StageName: constants.StageSlotSelection,
		Message:   "re-pick requested",
	}); err != nil {
		return models.Job{}, DatabaseFailure("failed to write re-pick log", map[string]any{
			"job_id": jobID,
		}, err)
	}

	if err := tx.Commit(); err != nil {
		return models.Job{}, DatabaseFailure("failed to commit re-pick", map[string]any{
			"job_id": jobID,
		}, err)
	}

	return sanitizeJob(job), nil
}

func (s *JobService) ProcessPendingAnalysis(ctx context.Context) error {
	submissionJobs, err := s.jobsRepository.ListByStatusAndStage(ctx, constants.JobStatusAnalyzing, constants.StageAnalysisSubmission)
	if err != nil {
		return err
	}
	for _, job := range submissionJobs {
		if err := s.processAnalysisSubmission(ctx, job); err != nil {
			return err
		}
	}

	pollJobs, err := s.jobsRepository.ListByStatusAndStage(ctx, constants.JobStatusAnalyzing, constants.StageAnalysisPoll)
	if err != nil {
		return err
	}
	for _, job := range pollJobs {
		if err := s.processAnalysisPoll(ctx, job); err != nil {
			return err
		}
	}

	generationSubmissionJobs, err := s.jobsRepository.ListByStatusAndStage(ctx, constants.JobStatusGenerating, constants.StageGenerationSubmit)
	if err != nil {
		return err
	}
	for _, job := range generationSubmissionJobs {
		if err := s.processGenerationSubmission(ctx, job); err != nil {
			return err
		}
	}

	generationPollJobs, err := s.jobsRepository.ListByStatusAndStage(ctx, constants.JobStatusGenerating, constants.StageGenerationPoll)
	if err != nil {
		return err
	}
	for _, job := range generationPollJobs {
		if err := s.processGenerationPoll(ctx, job); err != nil {
			return err
		}
	}

	renderSubmissionJobs, err := s.jobsRepository.ListByStatusAndStage(ctx, constants.JobStatusStitching, constants.StageRenderSubmit)
	if err != nil {
		return err
	}
	for _, job := range renderSubmissionJobs {
		if err := s.processRenderSubmission(ctx, job); err != nil {
			return err
		}
	}

	renderPollJobs, err := s.jobsRepository.ListByStatusAndStage(ctx, constants.JobStatusStitching, constants.StageRenderPoll)
	if err != nil {
		return err
	}
	for _, job := range renderPollJobs {
		if err := s.processRenderPoll(ctx, job); err != nil {
			return err
		}
	}

	return nil
}

func (s *JobService) processAnalysisSubmission(ctx context.Context, job models.Job) error {
	job.Metadata = ensureJobMetadata(job.Metadata)
	if requestID := metadataString(job.Metadata, internalAnalysisRequestIDKey); requestID != "" {
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

	response, err := s.analysisClient.SubmitAnalysis(ctx, AnalysisRequest{
		JobID:      job.ID,
		VideoPath:  campaign.VideoPath,
		ProductID:  product.ID,
		CampaignID: campaign.ID,
	})
	if err != nil {
		return s.failAnalysisJob(ctx, job, constants.StageAnalysisSubmission, "analysis submission failed")
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
	current.Metadata[internalAnalysisRequestIDKey] = response.RequestID
	current.Status = constants.JobStatusAnalyzing
	current.CurrentStage = constants.StageAnalysisPoll
	current.ProgressPercent = 20
	if err := jobRepo.UpdateState(ctx, current); err != nil {
		return err
	}
	if err := logRepo.Insert(ctx, models.JobLog{
		JobID:     job.ID,
		Timestamp: TimestampNow(),
		EventType: "stage_started",
		StageName: constants.StageAnalysisPoll,
		Message:   "submitted video for cloud analysis",
	}); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *JobService) processAnalysisPoll(ctx context.Context, job models.Job) error {
	job.Metadata = ensureJobMetadata(job.Metadata)
	requestID := metadataString(job.Metadata, internalAnalysisRequestIDKey)
	if requestID == "" {
		return s.failAnalysisJob(ctx, job, constants.StageAnalysisPoll, "analysis request is missing")
	}

	campaign, err := s.campaignRepository.GetByID(ctx, job.CampaignID)
	if err != nil {
		return err
	}
	product, err := s.productsRepository.GetByID(ctx, campaign.ProductID)
	if err != nil {
		return err
	}

	response, err := s.analysisClient.PollAnalysis(ctx, AnalysisPollRequest{
		JobID:      job.ID,
		RequestID:  requestID,
		VideoPath:  campaign.VideoPath,
		ProductID:  product.ID,
		CampaignID: campaign.ID,
		SourceFPS:  metadataFloat(job.Metadata, "source_fps"),
	})
	if err != nil {
		return s.failAnalysisJob(ctx, job, constants.StageAnalysisPoll, "analysis polling failed")
	}

	switch strings.ToLower(response.Status) {
	case "", "pending", "running", "processing":
		return nil
	case "failed", "error":
		return s.failAnalysisJob(ctx, job, constants.StageAnalysisPoll, "analysis failed")
	case "completed", "succeeded":
		normalizedScenes := normalizeScenes(job.ID, response.Scenes)
		slots, rankingRequestID, rankErr := s.rankSlotsWithOpenAI(ctx, job.ID, metadataFloat(job.Metadata, "source_fps"), product, normalizedScenes, stringSliceMetadata(job.Metadata, "rejected_slot_ids"))
		if rankErr != nil {
			return s.failAnalysisJob(ctx, job, constants.StageAnalysisPoll, "slot ranking failed")
		}
		return s.persistCompletedAnalysis(ctx, job, normalizedScenes, slots, response, rankingRequestID)
	default:
		return s.failAnalysisJob(ctx, job, constants.StageAnalysisPoll, fmt.Sprintf("analysis returned unknown status %q", response.Status))
	}
}

func (s *JobService) persistCompletedAnalysis(ctx context.Context, job models.Job, scenes []models.Scene, slots []models.Slot, response AnalysisPollResponse, rankingRequestID string) error {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	jobRepo := db.NewJobsRepository(tx)
	sceneRepo := db.NewScenesRepository(tx)
	slotRepo := db.NewSlotsRepository(tx)
	logRepo := db.NewJobLogsRepository(tx)

	current, err := jobRepo.GetByID(ctx, job.ID)
	if err != nil {
		return err
	}
	current.Metadata = ensureJobMetadata(current.Metadata)
	if response.PayloadRef != "" {
		current.Metadata[internalAnalysisPayloadRef] = response.PayloadRef
	}
	if rankingRequestID != "" {
		current.Metadata[internalSlotRankingRequestIDKey] = rankingRequestID
	}

	if err := sceneRepo.ReplaceForJob(ctx, current.ID, scenes); err != nil {
		return err
	}

	if len(slots) == 0 {
		current.Metadata["top_slot_ids"] = []string{}
		errorCode := constants.ErrorCodeNoSuitableSlot
		errorMessage := "no suitable slot found"
		completedAt := TimestampNow()
		current.Status = constants.JobStatusFailed
		current.CurrentStage = constants.StageSlotSelection
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
			StageName: constants.StageSlotSelection,
			Message:   "no suitable slot found",
		}); err != nil {
			return err
		}
		return tx.Commit()
	}

	if err := slotRepo.ReplaceForJob(ctx, current.ID, slots); err != nil {
		return err
	}

	current.Status = constants.JobStatusAnalyzing
	current.CurrentStage = constants.StageSlotSelection
	current.ProgressPercent = 40
	current.ErrorCode = nil
	current.ErrorMessage = nil
	current.CompletedAt = nil
	current.Metadata["top_slot_ids"] = slotIDs(slots)
	if err := jobRepo.UpdateState(ctx, current); err != nil {
		return err
	}
	if err := logRepo.Insert(ctx, models.JobLog{
		JobID:     current.ID,
		Timestamp: TimestampNow(),
		EventType: "stage_completed",
		StageName: constants.StageSlotSelection,
		Message:   "analysis complete and slots proposed",
	}); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *JobService) failAnalysisJob(ctx context.Context, job models.Job, stage, message string) error {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	jobRepo := db.NewJobsRepository(tx)
	logRepo := db.NewJobLogsRepository(tx)

	current, err := jobRepo.GetByID(ctx, job.ID)
	if err != nil {
		return err
	}

	errorCode := constants.ErrorCodeAnalysisFailed
	completedAt := TimestampNow()
	current.Status = constants.JobStatusFailed
	current.CurrentStage = stage
	current.ErrorCode = &errorCode
	errorMessage := message
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
		Message:   message,
	}); err != nil {
		return err
	}

	return tx.Commit()
}

func normalizeScenes(jobID string, scenes []models.Scene) []models.Scene {
	normalized := make([]models.Scene, 0, len(scenes))
	now := TimestampNow()
	for index, scene := range scenes {
		if scene.SceneNumber == 0 {
			scene.SceneNumber = index + 1
		}
		if scene.ID == "" {
			scene.ID = fmt.Sprintf("scene_%s_%03d", jobID, scene.SceneNumber)
		}
		scene.JobID = jobID
		if scene.CreatedAt == "" {
			scene.CreatedAt = now
		}
		normalized = append(normalized, scene)
	}
	return normalized
}

func rankSlots(jobID string, sourceFPS float64, product models.Product, scenes []models.Scene, rejectedSlotIDs []string) []models.Slot {
	rejectedSet := make(map[string]struct{}, len(rejectedSlotIDs))
	for _, id := range rejectedSlotIDs {
		rejectedSet[id] = struct{}{}
	}

	slots := make([]models.Slot, 0)
	now := TimestampNow()
	for _, scene := range scenes {
		slot, ok := rankSceneCandidate(jobID, sourceFPS, product, scene, now)
		if !ok {
			continue
		}
		if _, rejected := rejectedSet[slot.ID]; rejected {
			continue
		}
		slots = append(slots, slot)
	}

	slices.SortFunc(slots, func(left, right models.Slot) int {
		if left.Score > right.Score {
			return -1
		}
		if left.Score < right.Score {
			return 1
		}
		if left.AnchorStartFrame < right.AnchorStartFrame {
			return -1
		}
		if left.AnchorStartFrame > right.AnchorStartFrame {
			return 1
		}
		return strings.Compare(left.ID, right.ID)
	})

	if len(slots) > 3 {
		slots = slots[:3]
	}
	for index := range slots {
		slots[index].Rank = index + 1
	}
	return slots
}

func rankSceneCandidate(jobID string, sourceFPS float64, product models.Product, scene models.Scene, timestamp string) (models.Slot, bool) {
	motionScore := clamp01(floatValue(scene.MotionScore, 0.2))
	stabilityScore := clamp01(floatValue(scene.StabilityScore, 0.6))
	dialogueScore := clamp01(floatValue(scene.DialogueActivityScore, 0.3))
	quietWindowSeconds := floatValue(scene.LongestQuietWindowSeconds, 0)
	actionIntensity := clamp01(floatValue(scene.ActionIntensityScore, motionScore))
	abruptCutRisk := clamp01(floatValue(scene.AbruptCutRisk, 0.1))
	sceneDuration := scene.EndSeconds - scene.StartSeconds

	if motionScore > 0.65 || actionIntensity > 0.70 || abruptCutRisk > 0.70 || sceneDuration < 10 || quietWindowSeconds < 3 {
		return models.Slot{}, false
	}

	if sourceFPS <= 0 {
		sourceFPS = 24
	}
	insertionFrame := scene.StartFrame + int(math.Round((sceneDuration*sourceFPS)/2))
	if insertionFrame < scene.StartFrame {
		insertionFrame = scene.StartFrame
	}
	if insertionFrame >= scene.EndFrame {
		insertionFrame = scene.EndFrame - 1
	}
	if insertionFrame < scene.StartFrame {
		insertionFrame = scene.StartFrame
	}
	anchorStartFrame := insertionFrame
	anchorEndFrame := insertionFrame + 1
	if anchorEndFrame > scene.EndFrame {
		anchorEndFrame = scene.EndFrame
	}
	if anchorEndFrame <= anchorStartFrame {
		anchorEndFrame = anchorStartFrame + 1
	}

	quietWindowScore := clamp01(quietWindowSeconds / 3)
	contextRelevanceScore := scoreContextRelevance(product, scene)
	narrativeFitScore := clamp01((1-dialogueScore)*0.45 + quietWindowScore*0.25 + contextRelevanceScore*0.20 + (1-actionIntensity)*0.10)
	anchorContinuityScore := clamp01(stabilityScore*0.40 + (1-motionScore)*0.35 + (1-abruptCutRisk)*0.25)
	slotScore := stabilityScore*0.30 + quietWindowScore*0.25 + contextRelevanceScore*0.20 + narrativeFitScore*0.15 + anchorContinuityScore*0.10

	slotID := fmt.Sprintf("slot_%s_%s_%d_%d", jobID, scene.ID, anchorStartFrame, anchorEndFrame)
	reasoning := fmt.Sprintf(
		"low motion %.2f, %.1f-second quiet window, context match %.2f, continuity %.2f",
		motionScore,
		quietWindowSeconds,
		contextRelevanceScore,
		anchorContinuityScore,
	)

	return models.Slot{
		ID:                    slotID,
		JobID:                 jobID,
		SceneID:               scene.ID,
		AnchorStartFrame:      anchorStartFrame,
		AnchorEndFrame:        anchorEndFrame,
		QuietWindowSeconds:    quietWindowSeconds,
		Score:                 roundScore(slotScore),
		Reasoning:             reasoning,
		Status:                constants.SlotStatusProposed,
		ContextRelevanceScore: floatPtr(roundScore(contextRelevanceScore)),
		NarrativeFitScore:     floatPtr(roundScore(narrativeFitScore)),
		AnchorContinuityScore: floatPtr(roundScore(anchorContinuityScore)),
		CreatedAt:             timestamp,
		UpdatedAt:             timestamp,
	}, true
}

func scoreContextRelevance(product models.Product, scene models.Scene) float64 {
	weightedProductTerms := make(map[string]float64)
	addWeightedTerms := func(text string, weight float64) {
		for _, token := range tokenize(text) {
			weightedProductTerms[token] += weight
		}
	}

	addWeightedTerms(product.Name, 2)
	addWeightedTerms(product.Description, 1.5)
	addWeightedTerms(product.Category, 1)
	addWeightedTerms(product.SourceURL, 0.5)
	for _, keyword := range product.ContextKeywords {
		addWeightedTerms(keyword, 1.5)
	}
	if len(weightedProductTerms) == 0 {
		return 0
	}

	sceneTerms := make(map[string]struct{})
	for _, keyword := range scene.ContextKeywords {
		for _, token := range tokenize(keyword) {
			sceneTerms[token] = struct{}{}
		}
	}
	for _, token := range tokenize(scene.NarrativeSummary) {
		sceneTerms[token] = struct{}{}
	}

	var totalWeight float64
	var matchedWeight float64
	for token, weight := range weightedProductTerms {
		totalWeight += weight
		if _, ok := sceneTerms[token]; ok {
			matchedWeight += weight
		}
	}
	if totalWeight == 0 {
		return 0
	}
	return clamp01(matchedWeight / totalWeight)
}

func sanitizeJob(job models.Job) models.Job {
	job.Metadata = cloneMetadata(job.Metadata)
	delete(job.Metadata, internalAnalysisRequestIDKey)
	delete(job.Metadata, internalAnalysisPayloadRef)
	delete(job.Metadata, internalSlotRankingRequestIDKey)
	delete(job.Metadata, internalProductLineRequestIDKey)
	delete(job.Metadata, internalGenerationBriefRequestIDKey)
	delete(job.Metadata, internalGenerationRequestIDKey)
	delete(job.Metadata, internalGenerationPayloadRef)
	delete(job.Metadata, internalRenderRequestIDKey)
	delete(job.Metadata, internalRenderPayloadRef)
	job.Metadata = ensureJobMetadata(job.Metadata)
	return job
}

func sanitizePreview(preview models.Preview) models.Preview {
	if preview.OutputVideoPath != "" {
		preview.DownloadPath = fmt.Sprintf("/api/jobs/%s/preview/download", preview.JobID)
	}
	return preview
}

func ensureJobMetadata(metadata models.Metadata) models.Metadata {
	clean := cloneMetadata(metadata)
	if clean == nil {
		clean = models.Metadata{}
	}
	if _, ok := clean["rejected_slot_ids"]; !ok {
		clean["rejected_slot_ids"] = []string{}
	}
	if _, ok := clean["top_slot_ids"]; !ok {
		clean["top_slot_ids"] = []string{}
	}
	if _, ok := clean["repick_count"]; !ok {
		clean["repick_count"] = 0
	}
	return clean
}

func cloneMetadata(metadata models.Metadata) models.Metadata {
	if metadata == nil {
		return nil
	}
	cloned := make(models.Metadata, len(metadata))
	for key, value := range metadata {
		cloned[key] = value
	}
	return cloned
}

func metadataRepickCount(metadata models.Metadata) int {
	switch value := metadata["repick_count"].(type) {
	case int:
		return value
	case int32:
		return int(value)
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return 0
	}
}

func metadataFloat(metadata models.Metadata, key string) float64 {
	switch value := metadata[key].(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	default:
		return 0
	}
}

func metadataString(metadata models.Metadata, key string) string {
	value, ok := metadata[key].(string)
	if !ok {
		return ""
	}
	return value
}

func stringSliceMetadata(metadata models.Metadata, key string) []string {
	value, ok := metadata[key]
	if !ok || value == nil {
		return []string{}
	}
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []any:
		results := make([]string, 0, len(typed))
		for _, item := range typed {
			if str, ok := item.(string); ok && str != "" {
				results = append(results, str)
			}
		}
		return results
	default:
		return []string{}
	}
}

func proposedSlotIDs(slots []models.Slot) []string {
	ids := make([]string, 0, len(slots))
	for _, slot := range slots {
		if slot.Status == constants.SlotStatusProposed {
			ids = append(ids, slot.ID)
		}
	}
	return ids
}

func slotIDs(slots []models.Slot) []string {
	ids := make([]string, 0, len(slots))
	for _, slot := range slots {
		ids = append(ids, slot.ID)
	}
	return ids
}

func canRepick(slots []models.Slot, repickCount int) bool {
	if repickCount == 1 && len(slots) == 0 {
		return true
	}
	if len(slots) == 0 {
		return false
	}
	for _, slot := range slots {
		if slot.Status == constants.SlotStatusProposed || slot.Status == constants.SlotStatusSelected {
			return false
		}
	}
	return true
}

func appendUnique(values []string, candidate string) []string {
	if candidate == "" {
		return values
	}
	for _, value := range values {
		if value == candidate {
			return values
		}
	}
	return append(values, candidate)
}

func tokenize(value string) []string {
	replacer := strings.NewReplacer(",", " ", ".", " ", "/", " ", "-", " ", "_", " ", ":", " ", ";", " ", "(", " ", ")", " ")
	normalized := replacer.Replace(strings.ToLower(value))
	parts := strings.Fields(normalized)
	results := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(part) < 2 {
			continue
		}
		results = append(results, part)
	}
	return results
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func floatValue(value *float64, fallback float64) float64 {
	if value == nil {
		return fallback
	}
	return *value
}

func floatPtr(value float64) *float64 {
	return &value
}

func roundScore(value float64) float64 {
	return math.Round(value*1000) / 1000
}
