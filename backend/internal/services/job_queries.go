package services

import (
	"context"
	"errors"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/db"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

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

func (s *JobService) ListRecent(ctx context.Context, limit int) ([]models.Job, error) {
	if limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}

	jobs, err := s.jobsRepository.ListRecent(ctx, limit)
	if err != nil {
		return nil, DatabaseFailure("failed to list jobs", map[string]any{
			"limit": limit,
		}, err)
	}

	results := make([]models.Job, 0, len(jobs))
	for _, job := range jobs {
		results = append(results, sanitizeJob(job))
	}
	return results, nil
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
