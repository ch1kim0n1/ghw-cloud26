package services

import "github.com/ch1kim0n1/ghw-cloud26/backend/internal/db"

type CampaignService struct {
	repository *db.CampaignsRepository
}

func NewCampaignService(repository *db.CampaignsRepository) *CampaignService {
	return &CampaignService{repository: repository}
}
