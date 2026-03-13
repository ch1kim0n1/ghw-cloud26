package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

type JobsRepository struct {
	db dbExecutor
}

func NewJobsRepository(db dbExecutor) *JobsRepository {
	return &JobsRepository{db: db}
}

func (r *JobsRepository) Insert(ctx context.Context, job models.Job) error {
	metadataJSON, err := json.Marshal(job.Metadata)
	if err != nil {
		return fmt.Errorf("marshal job metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO jobs (
			id, campaign_id, status, current_stage, progress_percent, selected_slot_id, repick_count, error_code, error_message, metadata_json, created_at, started_at, completed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		job.ID,
		job.CampaignID,
		job.Status,
		nullIfEmpty(job.CurrentStage),
		job.ProgressPercent,
		job.SelectedSlotID,
		0,
		job.ErrorCode,
		job.ErrorMessage,
		string(metadataJSON),
		job.CreatedAt,
		job.StartedAt,
		job.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("insert job %s: %w", job.ID, err)
	}

	return nil
}
