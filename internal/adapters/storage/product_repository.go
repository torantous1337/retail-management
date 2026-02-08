package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/torantous1337/retail-management/internal/core/domain"
)

// ProductRepository implements the product repository using SQLite.
type ProductRepository struct {
	db *sqlx.DB
}

// NewProductRepository creates a new product repository instance.
func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// productRow is a database row representation for products.
type productRow struct {
	ID         string         `db:"id"`
	Name       string         `db:"name"`
	SKU        string         `db:"sku"`
	BasePrice  float64        `db:"base_price"`
	Properties sql.NullString `db:"properties"`
	CreatedAt  time.Time      `db:"created_at"`
	UpdatedAt  time.Time      `db:"updated_at"`
}

// Create creates a new product in the database.
func (r *ProductRepository) Create(ctx context.Context, product *domain.Product) error {
	// Serialize properties to JSON
	propertiesJSON, err := json.Marshal(product.Properties)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO products (id, name, sku, base_price, properties, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(ctx, query,
		product.ID,
		product.Name,
		product.SKU,
		product.BasePrice,
		string(propertiesJSON),
		product.CreatedAt,
		product.UpdatedAt,
	)

	return err
}

// GetByID retrieves a product by its ID.
func (r *ProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	query := `SELECT * FROM products WHERE id = ?`

	var row productRow
	err := r.db.GetContext(ctx, &row, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}

	return r.toDomain(&row)
}

// GetBySKU retrieves a product by its SKU.
func (r *ProductRepository) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	query := `SELECT * FROM products WHERE sku = ?`

	var row productRow
	err := r.db.GetContext(ctx, &row, query, sku)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}

	return r.toDomain(&row)
}

// List retrieves all products with pagination.
func (r *ProductRepository) List(ctx context.Context, limit, offset int) ([]*domain.Product, error) {
	query := `SELECT * FROM products ORDER BY created_at DESC LIMIT ? OFFSET ?`

	var rows []productRow
	err := r.db.SelectContext(ctx, &rows, query, limit, offset)
	if err != nil {
		return nil, err
	}

	products := make([]*domain.Product, 0, len(rows))
	for _, row := range rows {
		product, err := r.toDomain(&row)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}

// Update updates an existing product.
func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	// Serialize properties to JSON
	propertiesJSON, err := json.Marshal(product.Properties)
	if err != nil {
		return err
	}

	query := `
		UPDATE products
		SET name = ?, sku = ?, base_price = ?, properties = ?
		WHERE id = ?
	`

	_, err = r.db.ExecContext(ctx, query,
		product.Name,
		product.SKU,
		product.BasePrice,
		string(propertiesJSON),
		product.ID,
	)

	return err
}

// Delete deletes a product by its ID.
func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM products WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// toDomain converts a database row to a domain entity.
func (r *ProductRepository) toDomain(row *productRow) (*domain.Product, error) {
	product := &domain.Product{
		ID:        row.ID,
		Name:      row.Name,
		SKU:       row.SKU,
		BasePrice: row.BasePrice,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}

	// Deserialize properties from JSON
	if row.Properties.Valid && row.Properties.String != "" {
		var properties map[string]interface{}
		err := json.Unmarshal([]byte(row.Properties.String), &properties)
		if err != nil {
			return nil, err
		}
		product.Properties = properties
	} else {
		product.Properties = make(map[string]interface{})
	}

	return product, nil
}
