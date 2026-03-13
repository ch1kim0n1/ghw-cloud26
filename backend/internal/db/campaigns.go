package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

type CampaignsRepository struct {
	db dbExecutor
}

func NewCampaignsRepository(db dbExecutor) *CampaignsRepository {
	return &CampaignsRepository{db: db}
}

func (r *CampaignsRepository) Insert(ctx context.Context, campaign models.Campaign) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO campaigns (
			id, product_id, name, video_filename, video_path, source_fps, duration_seconds, target_ad_duration_seconds, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		campaign.ID,
		campaign.ProductID,
		campaign.Name,
		campaign.VideoFilename,
		campaign.VideoPath,
		campaign.SourceFPS,
		campaign.DurationSeconds,
		campaign.TargetAdDurationSeconds,
		campaign.CreatedAt,
		campaign.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert campaign %s: %w", campaign.ID, err)
	}
	return nil
}

func (r *CampaignsRepository) GetByID(ctx context.Context, campaignID string) (models.Campaign, error) {
	var campaign models.Campaign
	var (
		jobID        sql.NullString
		status       sql.NullString
		currentStage sql.NullString
		updatedAt    sql.NullString
	)

	err := r.db.QueryRowContext(ctx, `
		SELECT
			c.id,
			c.product_id,
			c.name,
			c.video_filename,
			c.video_path,
			c.source_fps,
			c.duration_seconds,
			c.target_ad_duration_seconds,
			c.created_at,
			c.updated_at,
			j.id,
			j.status,
			j.current_stage
		FROM campaigns c
		LEFT JOIN jobs j ON j.campaign_id = c.id
		WHERE c.id = ?
		LIMIT 1
	`, campaignID).Scan(
		&campaign.ID,
		&campaign.ProductID,
		&campaign.Name,
		&campaign.VideoFilename,
		&campaign.VideoPath,
		&campaign.SourceFPS,
		&campaign.DurationSeconds,
		&campaign.TargetAdDurationSeconds,
		&campaign.CreatedAt,
		&updatedAt,
		&jobID,
		&status,
		&currentStage,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Campaign{}, ErrNotFound
	}
	if err != nil {
		return models.Campaign{}, fmt.Errorf("query campaign %s: %w", campaignID, err)
	}

	campaign.UpdatedAt = updatedAt.String
	campaign.JobID = jobID.String
	campaign.Status = status.String
	campaign.CurrentStage = currentStage.String
	return campaign, nil
}
