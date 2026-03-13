package db

import "database/sql"

type CampaignsRepository struct {
	db *sql.DB
}

func NewCampaignsRepository(db *sql.DB) *CampaignsRepository {
	return &CampaignsRepository{db: db}
}
