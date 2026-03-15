package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

type WebsiteAdsRepository struct {
	db dbExecutor
}

func NewWebsiteAdsRepository(db dbExecutor) *WebsiteAdsRepository {
	return &WebsiteAdsRepository{db: db}
}

func (r *WebsiteAdsRepository) Insert(ctx context.Context, job models.WebsiteAdJob) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO website_ad_jobs (
			id, product_id, product_name, product_description, article_headline, article_body,
			brand_style, prompt, status, banner_image_path, vertical_image_path, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		job.ID,
		nullIfEmpty(job.ProductID),
		job.ProductName,
		nullIfEmpty(job.ProductDescription),
		job.ArticleHeadline,
		job.ArticleBody,
		nullIfEmpty(job.BrandStyle),
		job.Prompt,
		job.Status,
		nullIfEmpty(job.BannerImagePath),
		nullIfEmpty(job.VerticalImagePath),
		job.CreatedAt,
		job.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert website ad job %s: %w", job.ID, err)
	}
	return nil
}

func (r *WebsiteAdsRepository) UpdateResult(ctx context.Context, jobID, status, bannerPath, verticalPath, updatedAt string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE website_ad_jobs
		SET status = ?, banner_image_path = ?, vertical_image_path = ?, updated_at = ?
		WHERE id = ?
	`, status, nullIfEmpty(bannerPath), nullIfEmpty(verticalPath), updatedAt, jobID)
	if err != nil {
		return fmt.Errorf("update website ad job %s: %w", jobID, err)
	}
	return nil
}

func (r *WebsiteAdsRepository) GetByID(ctx context.Context, id string) (models.WebsiteAdJob, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, product_id, product_name, product_description, article_headline, article_body,
			brand_style, prompt, status, banner_image_path, vertical_image_path, created_at, updated_at
		FROM website_ad_jobs
		WHERE id = ?
	`, id)

	return scanWebsiteAdJob(row)
}

func (r *WebsiteAdsRepository) List(ctx context.Context) ([]models.WebsiteAdJob, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, product_id, product_name, product_description, article_headline, article_body,
			brand_style, prompt, status, banner_image_path, vertical_image_path, created_at, updated_at
		FROM website_ad_jobs
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list website ad jobs: %w", err)
	}
	defer rows.Close()

	jobs := make([]models.WebsiteAdJob, 0)
	for rows.Next() {
		job, err := scanWebsiteAdJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate website ad jobs: %w", err)
	}

	return jobs, nil
}

type websiteAdScanner interface {
	Scan(...any) error
}

func scanWebsiteAdJob(scanner websiteAdScanner) (models.WebsiteAdJob, error) {
	var (
		job                models.WebsiteAdJob
		productID          sql.NullString
		productDescription sql.NullString
		brandStyle         sql.NullString
		bannerPath         sql.NullString
		verticalPath       sql.NullString
	)

	err := scanner.Scan(
		&job.ID,
		&productID,
		&job.ProductName,
		&productDescription,
		&job.ArticleHeadline,
		&job.ArticleBody,
		&brandStyle,
		&job.Prompt,
		&job.Status,
		&bannerPath,
		&verticalPath,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.WebsiteAdJob{}, ErrNotFound
	}
	if err != nil {
		return models.WebsiteAdJob{}, fmt.Errorf("scan website ad job: %w", err)
	}

	job.ProductID = productID.String
	job.ProductDescription = productDescription.String
	job.BrandStyle = brandStyle.String
	job.BannerImagePath = bannerPath.String
	job.VerticalImagePath = verticalPath.String
	return job, nil
}
