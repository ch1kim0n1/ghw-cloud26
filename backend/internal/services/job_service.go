package services

import "github.com/ch1kim0n1/ghw-cloud26/backend/internal/db"

type JobService struct {
	repository *db.JobsRepository
}

func NewJobService(repository *db.JobsRepository) *JobService {
	return &JobService{repository: repository}
}
