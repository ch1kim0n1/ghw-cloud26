package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

type dbExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type ProductsRepository struct {
	db dbExecutor
}

func NewProductsRepository(db dbExecutor) *ProductsRepository {
	return &ProductsRepository{db: db}
}

func (r *ProductsRepository) Insert(ctx context.Context, product models.Product) error {
	keywordsJSON, err := json.Marshal(product.ContextKeywords)
	if err != nil {
		return fmt.Errorf("marshal product keywords: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO products (
			id, name, description, category, context_keywords_json, source_url, image_path, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		product.ID,
		product.Name,
		nullIfEmpty(product.Description),
		nullIfEmpty(product.Category),
		string(keywordsJSON),
		nullIfEmpty(product.SourceURL),
		nullIfEmpty(product.ImagePath),
		product.CreatedAt,
		product.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert product %s: %w", product.ID, err)
	}

	return nil
}

func (r *ProductsRepository) GetByID(ctx context.Context, id string) (models.Product, error) {
	var (
		product      models.Product
		keywordsJSON string
		description  sql.NullString
		category     sql.NullString
		sourceURL    sql.NullString
		imagePath    sql.NullString
		updatedAt    sql.NullString
	)

	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, description, category, context_keywords_json, source_url, image_path, created_at, updated_at
		FROM products
		WHERE id = ?
	`, id).Scan(
		&product.ID,
		&product.Name,
		&description,
		&category,
		&keywordsJSON,
		&sourceURL,
		&imagePath,
		&product.CreatedAt,
		&updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Product{}, ErrNotFound
	}
	if err != nil {
		return models.Product{}, fmt.Errorf("query product %s: %w", id, err)
	}

	product.Description = description.String
	product.Category = category.String
	product.SourceURL = sourceURL.String
	product.ImagePath = imagePath.String
	product.UpdatedAt = updatedAt.String

	if err := json.Unmarshal([]byte(keywordsJSON), &product.ContextKeywords); err != nil {
		return models.Product{}, fmt.Errorf("unmarshal product keywords %s: %w", id, err)
	}

	return product, nil
}

func (r *ProductsRepository) List(ctx context.Context) ([]models.Product, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, description, category, context_keywords_json, source_url, image_path, created_at, updated_at
		FROM products
		ORDER BY datetime(created_at) DESC, id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	products := make([]models.Product, 0)
	for rows.Next() {
		var (
			product      models.Product
			keywordsJSON string
			description  sql.NullString
			category     sql.NullString
			sourceURL    sql.NullString
			imagePath    sql.NullString
			updatedAt    sql.NullString
		)

		if err := rows.Scan(
			&product.ID,
			&product.Name,
			&description,
			&category,
			&keywordsJSON,
			&sourceURL,
			&imagePath,
			&product.CreatedAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan product row: %w", err)
		}

		product.Description = description.String
		product.Category = category.String
		product.SourceURL = sourceURL.String
		product.ImagePath = imagePath.String
		product.UpdatedAt = updatedAt.String

		if err := json.Unmarshal([]byte(keywordsJSON), &product.ContextKeywords); err != nil {
			return nil, fmt.Errorf("unmarshal product keywords %s: %w", product.ID, err)
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate product rows: %w", err)
	}

	return products, nil
}
