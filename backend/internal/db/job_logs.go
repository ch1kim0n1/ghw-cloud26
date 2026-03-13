package db

import "database/sql"

type JobLogsRepository struct {
	db *sql.DB
}

func NewJobLogsRepository(db *sql.DB) *JobLogsRepository {
	return &JobLogsRepository{db: db}
}
