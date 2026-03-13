package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/constants"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/db"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

type ProductService struct {
	repository *db.ProductsRepository
	storage    *LocalStorageService
	uploadDir  string
}

type CreateProductInput struct {
	Name            string
	Description     string
	Category        string
	ContextKeywords []string
	SourceURL       string
	ImageHeader     *multipart.FileHeader
	ImageFile       multipart.File
}

func NewProductService(repository *db.ProductsRepository, storage *LocalStorageService, uploadDir string) *ProductService {
	return &ProductService{
		repository: repository,
		storage:    storage,
		uploadDir:  uploadDir,
	}
}

func (s *ProductService) List(ctx context.Context) ([]models.Product, error) {
	products, err := s.repository.List(ctx)
	if err != nil {
		return nil, DatabaseFailure("failed to list products", nil, err)
	}
	return products, nil
}

func (s *ProductService) Create(ctx context.Context, input CreateProductInput) (models.Product, error) {
	if err := validateProductInput(input); err != nil {
		return models.Product{}, err
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

	if input.ImageHeader != nil && input.ImageFile != nil {
		imagePath, err := s.saveProductImage(product.ID, input.ImageHeader, input.ImageFile)
		if err != nil {
			return models.Product{}, err
		}
		product.ImagePath = imagePath
	}

	if err := s.repository.Insert(ctx, product); err != nil {
		if product.ImagePath != "" {
			_ = s.storage.Delete(product.ImagePath)
		}
		return models.Product{}, DatabaseFailure("failed to create product", map[string]any{
			"product_id": product.ID,
		}, err)
	}

	return product, nil
}

func validateProductInput(input CreateProductInput) error {
	name := strings.TrimSpace(input.Name)
	sourceURL := strings.TrimSpace(input.SourceURL)
	if name == "" {
		return InvalidRequest(constants.ErrorCodeInvalidRequest, "name is required", map[string]any{
			"field": "name",
		})
	}
	if sourceURL == "" && input.ImageHeader == nil {
		return InvalidRequest(constants.ErrorCodeProductInputMissing, "either source_url or image_file is required", nil)
	}
	if input.ImageHeader != nil {
		extension := strings.ToLower(filepath.Ext(input.ImageHeader.Filename))
		if extension != ".png" && extension != ".jpg" && extension != ".jpeg" {
			return InvalidRequest(constants.ErrorCodeInvalidRequest, "image_file must be PNG or JPG", map[string]any{
				"field": "image_file",
			})
		}
	}
	return nil
}

func (s *ProductService) saveProductImage(productID string, header *multipart.FileHeader, file multipart.File) (string, error) {
	extension := strings.ToLower(filepath.Ext(header.Filename))
	if extension == ".jpeg" {
		extension = ".jpg"
	}
	targetPath := filepath.Join(s.uploadDir, fmt.Sprintf("%s%s", productID, extension))
	if err := s.storage.SaveReader(targetPath, file); err != nil {
		return "", StorageFailure("failed to save product image", map[string]any{
			"product_id": productID,
			"path":       targetPath,
		}, err)
	}
	return targetPath, nil
}
