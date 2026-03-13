package db

import "database/sql"

type ProductsRepository struct {
	db *sql.DB
}

func NewProductsRepository(db *sql.DB) *ProductsRepository {
	return &ProductsRepository{db: db}
}
