package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/constants"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/db"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

type WebsiteAdService struct {
	repository    *db.WebsiteAdsRepository
	products      *db.ProductsRepository
	imageClient   WebsiteAdsImageClient
	storage       *LocalStorageService
	websiteAdsDir string
}

type CreateWebsiteAdInput struct {
	ProductID          string
	ProductName        string
	ProductDescription string
	ArticleHeadline    string
	ArticleBody        string
	BrandStyle         string
}

func NewWebsiteAdService(
	repository *db.WebsiteAdsRepository,
	products *db.ProductsRepository,
	imageClient WebsiteAdsImageClient,
	storage *LocalStorageService,
	websiteAdsDir string,
) *WebsiteAdService {
	return &WebsiteAdService{
		repository:    repository,
		products:      products,
		imageClient:   imageClient,
		storage:       storage,
		websiteAdsDir: websiteAdsDir,
	}
}

func (s *WebsiteAdService) List(ctx context.Context) ([]models.WebsiteAdJob, error) {
	jobs, err := s.repository.List(ctx)
	if err != nil {
		return nil, DatabaseFailure("failed to list website ad jobs", nil, err)
	}
	return jobs, nil
}

func (s *WebsiteAdService) GetByID(ctx context.Context, jobID string) (models.WebsiteAdJob, error) {
	job, err := s.repository.GetByID(ctx, strings.TrimSpace(jobID))
	if err == db.ErrNotFound {
		return models.WebsiteAdJob{}, ResourceNotFound("website ad job not found", map[string]any{"job_id": jobID}, err)
	}
	if err != nil {
		return models.WebsiteAdJob{}, DatabaseFailure("failed to load website ad job", map[string]any{"job_id": jobID}, err)
	}
	return job, nil
}

func (s *WebsiteAdService) Create(ctx context.Context, input CreateWebsiteAdInput) (models.WebsiteAdJob, error) {
	resolvedInput, err := s.resolveInput(ctx, input)
	if err != nil {
		return models.WebsiteAdJob{}, err
	}

	now := TimestampNow()
	job := models.WebsiteAdJob{
		ID:                 NewPrefixedID("wad"),
		ProductID:          resolvedInput.ProductID,
		ProductName:        resolvedInput.ProductName,
		ProductDescription: resolvedInput.ProductDescription,
		ArticleHeadline:    resolvedInput.ArticleHeadline,
		ArticleBody:        resolvedInput.ArticleBody,
		BrandStyle:         resolvedInput.BrandStyle,
		Prompt:             buildWebsiteAdPrompt(resolvedInput),
		Status:             constants.JobStatusGenerating,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := s.repository.Insert(ctx, job); err != nil {
		return models.WebsiteAdJob{}, DatabaseFailure("failed to create website ad job", map[string]any{"job_id": job.ID}, err)
	}

	banner, err := s.imageClient.Generate(ctx, WebsiteAdsImageGenerateRequest{
		Prompt: job.Prompt + " Compose it as a polished horizontal web banner layout.",
		Width:  1200,
		Height: 628,
	})
	if err != nil {
		_ = s.repository.UpdateResult(ctx, job.ID, constants.JobStatusFailed, "", "", TimestampNow())
		return models.WebsiteAdJob{}, providerFailure(err, job.ID)
	}

	vertical, err := s.imageClient.Generate(ctx, WebsiteAdsImageGenerateRequest{
		Prompt: job.Prompt + " Compose it as a tall vertical sidebar ad layout.",
		Width:  300,
		Height: 600,
	})
	if err != nil {
		_ = s.repository.UpdateResult(ctx, job.ID, constants.JobStatusFailed, "", "", TimestampNow())
		return models.WebsiteAdJob{}, providerFailure(err, job.ID)
	}

	bannerPath := filepath.Join(s.websiteAdsDir, fmt.Sprintf("%s_banner%s", job.ID, extensionFromContentType(banner.ContentType)))
	verticalPath := filepath.Join(s.websiteAdsDir, fmt.Sprintf("%s_vertical%s", job.ID, extensionFromContentType(vertical.ContentType)))

	if err := s.storage.Save(bannerPath, banner.Image); err != nil {
		_ = s.repository.UpdateResult(ctx, job.ID, constants.JobStatusFailed, "", "", TimestampNow())
		return models.WebsiteAdJob{}, StorageFailure("failed to save generated banner", map[string]any{"job_id": job.ID}, err)
	}
	if err := s.storage.Save(verticalPath, vertical.Image); err != nil {
		_ = s.storage.Delete(bannerPath)
		_ = s.repository.UpdateResult(ctx, job.ID, constants.JobStatusFailed, "", "", TimestampNow())
		return models.WebsiteAdJob{}, StorageFailure("failed to save generated vertical banner", map[string]any{"job_id": job.ID}, err)
	}

	job.Status = constants.JobStatusCompleted
	job.BannerImagePath = bannerPath
	job.VerticalImagePath = verticalPath
	job.UpdatedAt = TimestampNow()
	if err := s.repository.UpdateResult(ctx, job.ID, job.Status, bannerPath, verticalPath, job.UpdatedAt); err != nil {
		return models.WebsiteAdJob{}, DatabaseFailure("failed to finalize website ad job", map[string]any{"job_id": job.ID}, err)
	}

	return job, nil
}

func (s *WebsiteAdService) resolveInput(ctx context.Context, input CreateWebsiteAdInput) (CreateWebsiteAdInput, error) {
	resolved := CreateWebsiteAdInput{
		ProductID:          strings.TrimSpace(input.ProductID),
		ProductName:        strings.TrimSpace(input.ProductName),
		ProductDescription: strings.TrimSpace(input.ProductDescription),
		ArticleHeadline:    strings.TrimSpace(input.ArticleHeadline),
		ArticleBody:        strings.TrimSpace(input.ArticleBody),
		BrandStyle:         strings.TrimSpace(input.BrandStyle),
	}

	if resolved.ArticleHeadline == "" {
		return CreateWebsiteAdInput{}, InvalidRequest(constants.ErrorCodeInvalidRequest, "article_headline is required", map[string]any{"field": "article_headline"})
	}
	if resolved.ArticleBody == "" {
		return CreateWebsiteAdInput{}, InvalidRequest(constants.ErrorCodeInvalidRequest, "article_body is required", map[string]any{"field": "article_body"})
	}
	if len(resolved.ArticleBody) > 10000 {
		return CreateWebsiteAdInput{}, InvalidRequest(constants.ErrorCodeInvalidRequest, "article_body must be 10,000 characters or fewer", map[string]any{"field": "article_body"})
	}
	if resolved.ProductID != "" && resolved.ProductName != "" {
		return CreateWebsiteAdInput{}, InvalidRequest(constants.ErrorCodeInvalidRequest, "provide either product_id or inline product fields, not both", nil)
	}
	if resolved.ProductID == "" && resolved.ProductName == "" {
		return CreateWebsiteAdInput{}, InvalidRequest(constants.ErrorCodeProductInputMissing, "product_id or product_name is required", nil)
	}

	if resolved.ProductID != "" {
		product, err := s.products.GetByID(ctx, resolved.ProductID)
		if err == db.ErrNotFound {
			return CreateWebsiteAdInput{}, ResourceNotFound("product not found", map[string]any{"product_id": resolved.ProductID}, err)
		}
		if err != nil {
			return CreateWebsiteAdInput{}, DatabaseFailure("failed to load product", map[string]any{"product_id": resolved.ProductID}, err)
		}

		resolved.ProductName = product.Name
		if resolved.ProductDescription == "" {
			resolved.ProductDescription = product.Description
		}
	}

	if resolved.ProductName == "" {
		return CreateWebsiteAdInput{}, InvalidRequest(constants.ErrorCodeInvalidRequest, "product_name is required", map[string]any{"field": "product_name"})
	}
	if resolved.ProductDescription == "" {
		resolved.ProductDescription = "Product-forward web ad creative."
	}

	return resolved, nil
}

func buildWebsiteAdPrompt(input CreateWebsiteAdInput) string {
	headline := compactWhitespace(input.ArticleHeadline)
	body := truncateText(compactWhitespace(input.ArticleBody), 700)
	style := compactWhitespace(input.BrandStyle)
	if style == "" {
		style = "playful editorial"
	}

	return fmt.Sprintf(
		"Create a polished static website ad illustration that blends the article context with the advertised product. Article headline: %s. Article context: %s. Product: %s. Product details: %s. Visual style: %s, premium, vibrant, cohesive composition, product clearly visible, no readable text, no watermark, no collage, no split-screen, no brand logo.",
		headline,
		body,
		compactWhitespace(input.ProductName),
		compactWhitespace(input.ProductDescription),
		style,
	)
}

func compactWhitespace(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func truncateText(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return strings.TrimSpace(value[:limit]) + "..."
}

func extensionFromContentType(contentType string) string {
	switch {
	case strings.Contains(contentType, "png"):
		return ".png"
	case strings.Contains(contentType, "jpeg"), strings.Contains(contentType, "jpg"):
		return ".jpg"
	case strings.Contains(contentType, "webp"):
		return ".webp"
	default:
		return ".img"
	}
}

func providerFailure(err error, jobID string) error {
	return NewAppError(502, constants.ErrorCodeGenerationFailed, "website ad image generation failed", map[string]any{
		"job_id": jobID,
	}, err)
}
