package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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

func (r *JobsRepository) GetByID(ctx context.Context, id string) (models.Job, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			campaign_id,
			status,
			current_stage,
			progress_percent,
			selected_slot_id,
			repick_count,
			error_code,
			error_message,
			metadata_json,
			created_at,
			started_at,
			completed_at
		FROM jobs
		WHERE id = ?
		LIMIT 1
	`, id)

	job, err := scanJob(row)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Job{}, ErrNotFound
	}
	if err != nil {
		return models.Job{}, fmt.Errorf("query job %s: %w", id, err)
	}

	return job, nil
}

func (r *JobsRepository) ListByStatusAndStage(ctx context.Context, status, stage string) ([]models.Job, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			campaign_id,
			status,
			current_stage,
			progress_percent,
			selected_slot_id,
			repick_count,
			error_code,
			error_message,
			metadata_json,
			created_at,
			started_at,
			completed_at
		FROM jobs
		WHERE status = ? AND current_stage = ?
		ORDER BY datetime(created_at) ASC, id ASC
	`, status, stage)
	if err != nil {
		return nil, fmt.Errorf("list jobs by state: %w", err)
	}
	defer rows.Close()

	jobs := make([]models.Job, 0)
	for rows.Next() {
		job, scanErr := scanJob(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan job row: %w", scanErr)
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate jobs by state: %w", err)
	}

	return jobs, nil
}

func (r *JobsRepository) ListRecent(ctx context.Context, limit int) ([]models.Job, error) {
	if limit <= 0 {
		limit = 25
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			campaign_id,
			status,
			current_stage,
			progress_percent,
			selected_slot_id,
			repick_count,
			error_code,
			error_message,
			metadata_json,
			created_at,
			started_at,
			completed_at
		FROM jobs
		ORDER BY datetime(created_at) DESC, id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list recent jobs: %w", err)
	}
	defer rows.Close()

	jobs := make([]models.Job, 0, limit)
	for rows.Next() {
		job, scanErr := scanJob(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan recent job row: %w", scanErr)
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recent jobs: %w", err)
	}

	return jobs, nil
}

func (r *JobsRepository) UpdateState(ctx context.Context, job models.Job) error {
	metadataJSON, err := json.Marshal(job.Metadata)
	if err != nil {
		return fmt.Errorf("marshal updated job metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE jobs
		SET
			status = ?,
			current_stage = ?,
			progress_percent = ?,
			selected_slot_id = ?,
			repick_count = ?,
			error_code = ?,
			error_message = ?,
			metadata_json = ?,
			started_at = ?,
			completed_at = ?
		WHERE id = ?
	`,
		job.Status,
		nullIfEmpty(job.CurrentStage),
		job.ProgressPercent,
		job.SelectedSlotID,
		metadataRepickCount(job.Metadata),
		job.ErrorCode,
		job.ErrorMessage,
		string(metadataJSON),
		job.StartedAt,
		job.CompletedAt,
		job.ID,
	)
	if err != nil {
		return fmt.Errorf("update job %s: %w", job.ID, err)
	}

	return nil
}

type jobScanner interface {
	Scan(dest ...any) error
}

func scanJob(scanner jobScanner) (models.Job, error) {
	var (
		job            models.Job
		currentStage   sql.NullString
		selectedSlotID sql.NullString
		errorCode      sql.NullString
		errorMessage   sql.NullString
		metadataJSON   sql.NullString
		startedAt      sql.NullString
		completedAt    sql.NullString
		repickCount    int
	)

	err := scanner.Scan(
		&job.ID,
		&job.CampaignID,
		&job.Status,
		&currentStage,
		&job.ProgressPercent,
		&selectedSlotID,
		&repickCount,
		&errorCode,
		&errorMessage,
		&metadataJSON,
		&job.CreatedAt,
		&startedAt,
		&completedAt,
	)
	if err != nil {
		return models.Job{}, err
	}

	job.CurrentStage = currentStage.String
	if selectedSlotID.Valid {
		value := selectedSlotID.String
		job.SelectedSlotID = &value
	}
	if errorCode.Valid {
		value := errorCode.String
		job.ErrorCode = &value
	}
	if errorMessage.Valid {
		value := errorMessage.String
		job.ErrorMessage = &value
	}
	if startedAt.Valid {
		value := startedAt.String
		job.StartedAt = &value
	}
	if completedAt.Valid {
		value := completedAt.String
		job.CompletedAt = &value
	}

	job.Metadata = models.Metadata{}
	if metadataJSON.Valid && metadataJSON.String != "" {
		if err := json.Unmarshal([]byte(metadataJSON.String), &job.Metadata); err != nil {
			return models.Job{}, fmt.Errorf("unmarshal job metadata %s: %w", job.ID, err)
		}
	}
	if job.Metadata == nil {
		job.Metadata = models.Metadata{}
	}
	job.Metadata["repick_count"] = repickCount
	if _, ok := job.Metadata["rejected_slot_ids"]; !ok {
		job.Metadata["rejected_slot_ids"] = []string{}
	}
	if _, ok := job.Metadata["top_slot_ids"]; !ok {
		job.Metadata["top_slot_ids"] = []string{}
	}

	return job, nil
}

func metadataRepickCount(metadata models.Metadata) int {
	if metadata == nil {
		return 0
	}
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
