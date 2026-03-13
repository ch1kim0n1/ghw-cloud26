package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

type PreviewsRepository struct {
	db dbExecutor
}

func NewPreviewsRepository(db dbExecutor) *PreviewsRepository {
	return &PreviewsRepository{db: db}
}

func (r *PreviewsRepository) GetByJobID(ctx context.Context, jobID string) (models.Preview, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			job_id,
			slot_id,
			status,
			output_video_path,
			duration_seconds,
			render_retry_count,
			artifact_manifest_json,
			render_metrics_json,
			error_code,
			error_message,
			created_at,
			completed_at
		FROM job_previews
		WHERE job_id = ?
		LIMIT 1
	`, jobID)

	preview, err := scanPreview(row)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Preview{}, ErrNotFound
	}
	if err != nil {
		return models.Preview{}, fmt.Errorf("query preview for job %s: %w", jobID, err)
	}

	return preview, nil
}

func (r *PreviewsRepository) UpsertPending(ctx context.Context, preview models.Preview) error {
	artifactManifestJSON, err := json.Marshal(preview.ArtifactManifest)
	if err != nil {
		return fmt.Errorf("marshal preview artifact manifest: %w", err)
	}
	renderMetricsJSON, err := json.Marshal(preview.RenderMetrics)
	if err != nil {
		return fmt.Errorf("marshal preview render metrics: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO job_previews (
			id, job_id, slot_id, status, output_video_path, duration_seconds, render_retry_count, artifact_manifest_json, render_metrics_json, error_code, error_message, created_at, completed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(job_id) DO UPDATE SET
			slot_id = excluded.slot_id,
			status = excluded.status,
			output_video_path = NULL,
			duration_seconds = NULL,
			artifact_manifest_json = excluded.artifact_manifest_json,
			render_metrics_json = excluded.render_metrics_json,
			error_code = NULL,
			error_message = NULL,
			completed_at = NULL
	`,
		preview.ID,
		preview.JobID,
		preview.SlotID,
		preview.Status,
		nil,
		nil,
		preview.RenderRetryCount,
		string(artifactManifestJSON),
		string(renderMetricsJSON),
		nil,
		nil,
		preview.CreatedAt,
		nil,
	)
	if err != nil {
		return fmt.Errorf("upsert preview for job %s: %w", preview.JobID, err)
	}

	return nil
}

func (r *PreviewsRepository) MarkStitching(ctx context.Context, jobID, slotID string, artifactManifest models.Metadata) error {
	artifactManifestJSON, err := json.Marshal(artifactManifest)
	if err != nil {
		return fmt.Errorf("marshal preview artifact manifest: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE job_previews
		SET
			slot_id = ?,
			status = 'stitching',
			artifact_manifest_json = ?,
			error_code = NULL,
			error_message = NULL,
			completed_at = NULL
		WHERE job_id = ?
	`, slotID, string(artifactManifestJSON), jobID)
	if err != nil {
		return fmt.Errorf("mark preview stitching for job %s: %w", jobID, err)
	}
	return nil
}

func (r *PreviewsRepository) MarkCompleted(ctx context.Context, preview models.Preview) error {
	artifactManifestJSON, err := json.Marshal(preview.ArtifactManifest)
	if err != nil {
		return fmt.Errorf("marshal preview artifact manifest: %w", err)
	}
	renderMetricsJSON, err := json.Marshal(preview.RenderMetrics)
	if err != nil {
		return fmt.Errorf("marshal preview render metrics: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE job_previews
		SET
			slot_id = ?,
			status = 'completed',
			output_video_path = ?,
			duration_seconds = ?,
			artifact_manifest_json = ?,
			render_metrics_json = ?,
			error_code = NULL,
			error_message = NULL,
			completed_at = ?
		WHERE job_id = ?
	`,
		preview.SlotID,
		nullIfEmpty(preview.OutputVideoPath),
		nullableFloat(preview.DurationSeconds),
		string(artifactManifestJSON),
		string(renderMetricsJSON),
		preview.CompletedAt,
		preview.JobID,
	)
	if err != nil {
		return fmt.Errorf("mark preview completed for job %s: %w", preview.JobID, err)
	}
	return nil
}

func (r *PreviewsRepository) MarkFailed(ctx context.Context, jobID, slotID, errorCode, errorMessage, completedAt string, incrementRetry bool, artifactManifest models.Metadata) error {
	artifactManifestJSON, err := json.Marshal(artifactManifest)
	if err != nil {
		return fmt.Errorf("marshal preview artifact manifest: %w", err)
	}

	increment := 0
	if incrementRetry {
		increment = 1
	}
	_, err = r.db.ExecContext(ctx, `
		UPDATE job_previews
		SET
			slot_id = ?,
			status = 'failed',
			artifact_manifest_json = ?,
			error_code = ?,
			error_message = ?,
			completed_at = ?,
			render_retry_count = render_retry_count + ?
		WHERE job_id = ?
	`, slotID, string(artifactManifestJSON), nullIfEmpty(errorCode), nullIfEmpty(errorMessage), nullIfEmpty(completedAt), increment, jobID)
	if err != nil {
		return fmt.Errorf("mark preview failed for job %s: %w", jobID, err)
	}
	return nil
}

type previewScanner interface {
	Scan(dest ...any) error
}

func scanPreview(scanner previewScanner) (models.Preview, error) {
	var (
		preview              models.Preview
		outputVideoPath      sql.NullString
		durationSeconds      sql.NullFloat64
		artifactManifestJSON sql.NullString
		renderMetricsJSON    sql.NullString
		errorCode            sql.NullString
		errorMessage         sql.NullString
		completedAt          sql.NullString
	)

	err := scanner.Scan(
		&preview.ID,
		&preview.JobID,
		&preview.SlotID,
		&preview.Status,
		&outputVideoPath,
		&durationSeconds,
		&preview.RenderRetryCount,
		&artifactManifestJSON,
		&renderMetricsJSON,
		&errorCode,
		&errorMessage,
		&preview.CreatedAt,
		&completedAt,
	)
	if err != nil {
		return models.Preview{}, err
	}

	if outputVideoPath.Valid {
		preview.OutputVideoPath = outputVideoPath.String
	}
	if durationSeconds.Valid {
		preview.DurationSeconds = durationSeconds.Float64
	}
	if errorCode.Valid {
		value := errorCode.String
		preview.ErrorCode = &value
	}
	if errorMessage.Valid {
		value := errorMessage.String
		preview.ErrorMessage = &value
	}
	if completedAt.Valid {
		value := completedAt.String
		preview.CompletedAt = &value
	}

	preview.ArtifactManifest = models.Metadata{}
	if artifactManifestJSON.Valid && artifactManifestJSON.String != "" {
		if err := json.Unmarshal([]byte(artifactManifestJSON.String), &preview.ArtifactManifest); err != nil {
			return models.Preview{}, fmt.Errorf("unmarshal preview artifact manifest %s: %w", preview.ID, err)
		}
	}
	if preview.ArtifactManifest == nil {
		preview.ArtifactManifest = models.Metadata{}
	}

	preview.RenderMetrics = models.Metadata{}
	if renderMetricsJSON.Valid && renderMetricsJSON.String != "" {
		if err := json.Unmarshal([]byte(renderMetricsJSON.String), &preview.RenderMetrics); err != nil {
			return models.Preview{}, fmt.Errorf("unmarshal preview render metrics %s: %w", preview.ID, err)
		}
	}
	if preview.RenderMetrics == nil {
		preview.RenderMetrics = models.Metadata{}
	}

	return preview, nil
}

func nullableFloat(value float64) any {
	if value <= 0 {
		return nil
	}
	return value
}
