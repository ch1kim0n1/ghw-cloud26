package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

type JobLogsRepository struct {
	db dbExecutor
}

func NewJobLogsRepository(db dbExecutor) *JobLogsRepository {
	return &JobLogsRepository{db: db}
}

func (r *JobLogsRepository) Insert(ctx context.Context, logEntry models.JobLog) error {
	detailsJSON, err := json.Marshal(logEntry.Details)
	if err != nil {
		return fmt.Errorf("marshal job log details: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO job_logs (
			job_id, timestamp, event_type, stage_name, message, details_json
		) VALUES (?, ?, ?, ?, ?, ?)
	`,
		logEntry.JobID,
		logEntry.Timestamp,
		logEntry.EventType,
		nullIfEmpty(logEntry.StageName),
		logEntry.Message,
		string(detailsJSON),
	)
	if err != nil {
		return fmt.Errorf("insert job log for %s: %w", logEntry.JobID, err)
	}

	return nil
}

func (r *JobLogsRepository) ListByJobID(ctx context.Context, jobID string) ([]models.JobLog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, job_id, timestamp, event_type, stage_name, message, details_json
		FROM job_logs
		WHERE job_id = ?
		ORDER BY datetime(timestamp) ASC, id ASC
	`, jobID)
	if err != nil {
		return nil, fmt.Errorf("list job logs for %s: %w", jobID, err)
	}
	defer rows.Close()

	logs := make([]models.JobLog, 0)
	for rows.Next() {
		var (
			entry       models.JobLog
			stageName   sql.NullString
			detailsJSON sql.NullString
		)
		if err := rows.Scan(
			&entry.ID,
			&entry.JobID,
			&entry.Timestamp,
			&entry.EventType,
			&stageName,
			&entry.Message,
			&detailsJSON,
		); err != nil {
			return nil, fmt.Errorf("scan job log row: %w", err)
		}

		entry.StageName = stageName.String
		if detailsJSON.Valid && detailsJSON.String != "" {
			entry.Details = models.Metadata{}
			if err := json.Unmarshal([]byte(detailsJSON.String), &entry.Details); err != nil {
				return nil, fmt.Errorf("unmarshal job log details for %s: %w", entry.JobID, err)
			}
		}
		logs = append(logs, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate job log rows: %w", err)
	}

	return logs, nil
}
