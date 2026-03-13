package db

import "database/sql"

type SlotsRepository struct {
	db *sql.DB
}

func NewSlotsRepository(db *sql.DB) *SlotsRepository {
	return &SlotsRepository{db: db}
}
