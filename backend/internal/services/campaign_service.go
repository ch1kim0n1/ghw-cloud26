package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/constants"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/db"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

type CampaignService struct {
	database           *sql.DB
	repository         *db.CampaignsRepository
	productsRepository *db.ProductsRepository
	jobsRepository     *db.JobsRepository
	storage            *LocalStorageService
	inspector          *MediaInspector
	videoUploadDir     string
	imageUploadDir     string
}

type CreateCampaignInput struct {
	Name                    string
	TargetAdDurationSeconds string
	ProductID               string
	VideoHeader             *multipart.FileHeader
	VideoFile               multipart.File
	InlineProduct           CreateProductInput
}

func NewCampaignService(
	database *sql.DB,
	repository *db.CampaignsRepository,
	productsRepository *db.ProductsRepository,
	jobsRepository *db.JobsRepository,
	storage *LocalStorageService,
	inspector *MediaInspector,
	videoUploadDir string,
	imageUploadDir string,
) *CampaignService {
	return &CampaignService{
		database:           database,
		repository:         repository,
		productsRepository: productsRepository,
		jobsRepository:     jobsRepository,
		storage:            storage,
		inspector:          inspector,
		videoUploadDir:     videoUploadDir,
		imageUploadDir:     imageUploadDir,
	}
}

func (s *CampaignService) Get(ctx context.Context, campaignID string) (models.Campaign, error) {
	campaign, err := s.repository.GetByID(ctx, campaignID)
	if errors.Is(err, db.ErrNotFound) {
		return models.Campaign{}, ResourceNotFound("campaign not found", map[string]any{
			"campaign_id": campaignID,
		}, err)
	}
	if err != nil {
		return models.Campaign{}, DatabaseFailure("failed to load campaign", map[string]any{
			"campaign_id": campaignID,
		}, err)
	}
	return campaign, nil
}

func (s *CampaignService) Create(ctx context.Context, input CreateCampaignInput) (models.Campaign, error) {
	targetDuration, err := parseTargetAdDuration(input.TargetAdDurationSeconds)
	if err != nil {
		return models.Campaign{}, err
	}
	if err := validateCampaignProductMode(input); err != nil {
		return models.Campaign{}, err
	}
	if strings.TrimSpace(input.Name) == "" {
		return models.Campaign{}, InvalidRequest(constants.ErrorCodeInvalidRequest, "name is required", map[string]any{
			"field": "name",
		})
	}
	if input.VideoHeader == nil || input.VideoFile == nil {
		return models.Campaign{}, InvalidRequest(constants.ErrorCodeInvalidRequest, "video_file is required", map[string]any{
			"field": "video_file",
		})
	}

	campaignID := NewPrefixedID("camp")
	jobID := NewPrefixedID("job")
	videoPath := filepath.Join(s.videoUploadDir, fmt.Sprintf("%s%s", campaignID, strings.ToLower(filepath.Ext(input.VideoHeader.Filename))))

	if err := s.storage.SaveReader(videoPath, input.VideoFile); err != nil {
		return models.Campaign{}, StorageFailure("failed to save campaign video", map[string]any{
			"campaign_id": campaignID,
			"path":        videoPath,
		}, err)
	}

	mediaInfo, err := s.inspector.Inspect(ctx, videoPath)
	if err != nil {
		_ = s.storage.Delete(videoPath)
		return models.Campaign{}, InvalidRequest(constants.ErrorCodeInvalidVideoCodec, "failed to inspect uploaded video", map[string]any{
			"path": videoPath,
		})
	}
	if !strings.Contains(mediaInfo.FormatName, "mp4") || mediaInfo.VideoCodec != "h264" {
		_ = s.storage.Delete(videoPath)
		return models.Campaign{}, InvalidRequest(constants.ErrorCodeInvalidVideoCodec, "video_file must be an H.264 MP4", map[string]any{
			"codec":  mediaInfo.VideoCodec,
			"format": mediaInfo.FormatName,
		})
	}
	if mediaInfo.DurationSeconds < 600 || mediaInfo.DurationSeconds > 1200 {
		_ = s.storage.Delete(videoPath)
		return models.Campaign{}, InvalidRequest(constants.ErrorCodeInvalidVideoLength, "video duration must be between 10 and 20 minutes", map[string]any{
			"duration_seconds": mediaInfo.DurationSeconds,
		})
	}

	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		_ = s.storage.Delete(videoPath)
		return models.Campaign{}, DatabaseFailure("failed to create campaign transaction", nil, err)
	}

	var (
		imagePath string
		productID string
	)

	defer func() {
		if err != nil {
			_ = tx.Rollback()
			_ = s.storage.Delete(videoPath)
			if imagePath != "" {
				_ = s.storage.Delete(imagePath)
			}
		}
	}()

	productRepo := db.NewProductsRepository(tx)
	campaignRepo := db.NewCampaignsRepository(tx)
	jobRepo := db.NewJobsRepository(tx)

	if strings.TrimSpace(input.ProductID) != "" {
		productID = strings.TrimSpace(input.ProductID)
		if _, getErr := productRepo.GetByID(ctx, productID); getErr != nil {
			if errors.Is(getErr, db.ErrNotFound) {
				err = ResourceNotFound("product not found", map[string]any{
					"product_id": productID,
				}, getErr)
				return models.Campaign{}, err
			}
			err = DatabaseFailure("failed to load product", map[string]any{
				"product_id": productID,
			}, getErr)
			return models.Campaign{}, err
		}
	} else {
		var inlineProduct models.Product
		inlineProduct, imagePath, err = s.createInlineProduct(ctx, productRepo, input.InlineProduct)
		if err != nil {
			return models.Campaign{}, err
		}
		productID = inlineProduct.ID
	}

	now := TimestampNow()
	campaign := models.Campaign{
		ID:                      campaignID,
		JobID:                   jobID,
		ProductID:               productID,
		Name:                    strings.TrimSpace(input.Name),
		Status:                  constants.JobStatusQueued,
		CurrentStage:            constants.StageReadyForAnalysis,
		VideoFilename:           input.VideoHeader.Filename,
		VideoPath:               videoPath,
		SourceFPS:               mediaInfo.SourceFPS,
		DurationSeconds:         mediaInfo.DurationSeconds,
		TargetAdDurationSeconds: targetDuration,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	job := models.Job{
		ID:              jobID,
		CampaignID:      campaignID,
		Status:          constants.JobStatusQueued,
		CurrentStage:    constants.StageReadyForAnalysis,
		ProgressPercent: 0,
		CreatedAt:       now,
		Metadata: models.Metadata{
			"source_fps":        mediaInfo.SourceFPS,
			"duration_seconds":  mediaInfo.DurationSeconds,
			"rejected_slot_ids": []string{},
			"top_slot_ids":      []string{},
		},
	}

	if insertErr := campaignRepo.Insert(ctx, campaign); insertErr != nil {
		err = DatabaseFailure("failed to create campaign", map[string]any{
			"campaign_id": campaignID,
		}, insertErr)
		return models.Campaign{}, err
	}
	if insertErr := jobRepo.Insert(ctx, job); insertErr != nil {
		err = DatabaseFailure("failed to create job", map[string]any{
			"job_id": jobID,
		}, insertErr)
		return models.Campaign{}, err
	}
	if commitErr := tx.Commit(); commitErr != nil {
		err = DatabaseFailure("failed to commit campaign creation", map[string]any{
			"campaign_id": campaignID,
			"job_id":      jobID,
		}, commitErr)
		return models.Campaign{}, err
	}

	return campaign, nil
}

func (s *CampaignService) createInlineProduct(ctx context.Context, repository *db.ProductsRepository, input CreateProductInput) (models.Product, string, error) {
	if err := validateProductInput(input); err != nil {
		return models.Product{}, "", err
	}

	now := TimestampNow()
	product := models.Product{
		ID:              NewPrefixedID("prod"),
		Name:            strings.TrimSpace(input.Name),
		Description:     strings.TrimSpace(input.Description),
		Category:        strings.TrimSpace(input.Category),
		ContextKeywords: NormalizeKeywords(input.ContextKeywords),
		SourceURL:       strings.TrimSpace(input.SourceURL),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	imagePath := ""
	if input.ImageHeader != nil && input.ImageFile != nil {
		extension := strings.ToLower(filepath.Ext(input.ImageHeader.Filename))
		if extension == ".jpeg" {
			extension = ".jpg"
		}
		imagePath = filepath.Join(s.imageUploadDir, fmt.Sprintf("%s%s", product.ID, extension))
		if saveErr := s.storage.SaveReader(imagePath, input.ImageFile); saveErr != nil {
			return models.Product{}, "", StorageFailure("failed to save inline product image", map[string]any{
				"product_id": product.ID,
			}, saveErr)
		}
		product.ImagePath = imagePath
	}

	if insertErr := repository.Insert(ctx, product); insertErr != nil {
		return models.Product{}, imagePath, DatabaseFailure("failed to create inline product", map[string]any{
			"product_id": product.ID,
		}, insertErr)
	}

	return product, imagePath, nil
}

func parseTargetAdDuration(value string) (int, error) {
	if strings.TrimSpace(value) == "" {
		return 6, nil
	}
	duration, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, InvalidRequest(constants.ErrorCodeInvalidRequest, "target_ad_duration_seconds must be an integer", map[string]any{
			"field": "target_ad_duration_seconds",
		})
	}
	if duration <= 0 || duration > 8 {
		return 0, InvalidRequest(constants.ErrorCodeInvalidRequest, "target_ad_duration_seconds must be between 1 and 8", map[string]any{
			"field": "target_ad_duration_seconds",
		})
	}
	return duration, nil
}

func validateCampaignProductMode(input CreateCampaignInput) error {
	hasExisting := strings.TrimSpace(input.ProductID) != ""
	hasInline := hasInlineProductInput(input.InlineProduct)
	if hasExisting && hasInline {
		return InvalidRequest(constants.ErrorCodeInvalidRequest, "provide either product_id or inline product fields, not both", nil)
	}
	if !hasExisting && !hasInline {
		return InvalidRequest(constants.ErrorCodeProductInputMissing, "either product_id or inline product fields are required", nil)
	}
	return nil
}

func hasInlineProductInput(input CreateProductInput) bool {
	return strings.TrimSpace(input.Name) != "" ||
		strings.TrimSpace(input.Description) != "" ||
		strings.TrimSpace(input.Category) != "" ||
		len(NormalizeKeywords(input.ContextKeywords)) > 0 ||
		strings.TrimSpace(input.SourceURL) != "" ||
		input.ImageHeader != nil
}
