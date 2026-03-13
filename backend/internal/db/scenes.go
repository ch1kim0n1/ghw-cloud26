package db

import "database/sql"

type ScenesRepository struct {
	db *sql.DB
}

func NewScenesRepository(db *sql.DB) *ScenesRepository {
	return &ScenesRepository{db: db}
}
