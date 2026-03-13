package db

import "database/sql"

type JobsRepository struct {
	db *sql.DB
}

func NewJobsRepository(db *sql.DB) *JobsRepository {
	return &JobsRepository{db: db}
}
