package db

import "database/sql"

type PreviewsRepository struct {
	db *sql.DB
}

func NewPreviewsRepository(db *sql.DB) *PreviewsRepository {
	return &PreviewsRepository{db: db}
}
