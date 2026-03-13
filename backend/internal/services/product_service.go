package services

import "github.com/ch1kim0n1/ghw-cloud26/backend/internal/db"

type ProductService struct {
	repository *db.ProductsRepository
}

func NewProductService(repository *db.ProductsRepository) *ProductService {
	return &ProductService{repository: repository}
}
