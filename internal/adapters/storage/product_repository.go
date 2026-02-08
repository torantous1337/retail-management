package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/torantous1337/retail-management/internal/core/domain"
)

// ProductRepository implements the product repository using SQLite.
type ProductRepository struct {
	db sqlx.ExtContext
}

// NewProductRepository creates a new product repository instance.
func NewProductRepository(db sqlx.ExtContext) *ProductRepository {
	return &ProductRepository{db: db}
}

// productRow is a database row representation for products.
type productRow struct {
	ID         string         `db:"id"`
	Name       string         `db:"name"`
	SKU        string         `db:"sku"`
	CategoryID sql.NullString `db:"category_id"`
	BasePrice  float64        `db:"base_price"`
	Quantity   int            `db:"quantity"`
	CostPrice  float64        `db:"cost_price"`
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
		INSERT INTO products (id, name, sku, category_id, base_price, quantity, cost_price, properties, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(ctx, query,
		product.ID,
		product.Name,
		product.SKU,
		sql.NullString{String: product.CategoryID, Valid: product.CategoryID != ""},
		product.BasePrice,
		product.Quantity,
		product.CostPrice,
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
	err := sqlx.GetContext(ctx, r.db, &row, query, id)
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
	err := sqlx.GetContext(ctx, r.db, &row, query, sku)
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
	err := sqlx.SelectContext(ctx, r.db, &rows, query, limit, offset)
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

// validPropertyKey matches only safe JSON key names (alphanumeric + underscore).
var validPropertyKey = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

// Search retrieves products matching the given filter options.
// allowedKeys is the safelist of property keys from the category blueprint;
// any Properties filter key not in this list is silently ignored to prevent
// SQL injection via json_extract paths.
func (r *ProductRepository) Search(ctx context.Context, opts domain.FilterOptions, allowedKeys []string) ([]*domain.Product, error) {
	allowed := make(map[string]bool, len(allowedKeys))
	for _, k := range allowedKeys {
		allowed[k] = true
	}

	var clauses []string
	var args []interface{}

	// Full-text search via FTS5
	if opts.Query != "" {
		clauses = append(clauses, `p.rowid IN (SELECT rowid FROM products_fts WHERE products_fts MATCH ?)`)
		args = append(args, opts.Query)
	}

	// Category filter
	if opts.CategoryID != "" {
		clauses = append(clauses, `p.category_id = ?`)
		args = append(args, opts.CategoryID)
	}

	// Price range
	if opts.MinPrice != nil {
		clauses = append(clauses, `p.base_price >= ?`)
		args = append(args, *opts.MinPrice)
	}
	if opts.MaxPrice != nil {
		clauses = append(clauses, `p.base_price <= ?`)
		args = append(args, *opts.MaxPrice)
	}

	// Dynamic JSON property filters (safelisted keys only)
	for key, val := range opts.Properties {
		if !allowed[key] || !validPropertyKey.MatchString(key) {
			continue
		}
		clauses = append(clauses, fmt.Sprintf(`json_extract(p.properties, '$.%s') = ?`, key))
		args = append(args, val)
	}

	query := `SELECT p.* FROM products p`
	if len(clauses) > 0 {
		query += ` WHERE ` + strings.Join(clauses, " AND ")
	}
	query += ` ORDER BY p.created_at DESC`

	limit := opts.Limit
	if limit <= 0 {
		limit = 10
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}
	query += ` LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	var rows []productRow
	err := sqlx.SelectContext(ctx, r.db, &rows, query, args...)
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

// inventorySummaryRow holds a row from the inventory summary query.
type inventorySummaryRow struct {
	CategoryID   sql.NullString `db:"category_id"`
	CategoryName sql.NullString `db:"category_name"`
	Count        int            `db:"count"`
	TotalValue   float64        `db:"total_value"`
}

// GetInventorySummary returns aggregated inventory analytics.
func (r *ProductRepository) GetInventorySummary(ctx context.Context) (*domain.InventorySummary, error) {
	query := `
		SELECT
			p.category_id,
			c.name AS category_name,
			COUNT(*) AS count,
			COALESCE(SUM(p.base_price), 0) AS total_value
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		GROUP BY p.category_id
	`

	var rows []inventorySummaryRow
	err := sqlx.SelectContext(ctx, r.db, &rows, query)
	if err != nil {
		return nil, err
	}

	summary := &domain.InventorySummary{}
	for _, row := range rows {
		bd := domain.CategoryBreakdown{
			CategoryID:   row.CategoryID.String,
			CategoryName: row.CategoryName.String,
			Count:        row.Count,
			TotalValue:   row.TotalValue,
		}
		if !row.CategoryID.Valid {
			bd.CategoryName = "Uncategorized"
		}
		summary.CategoryBreakdown = append(summary.CategoryBreakdown, bd)
		summary.TotalItems += row.Count
		summary.TotalValue += row.TotalValue
	}

	return summary, nil
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
		SET name = ?, sku = ?, category_id = ?, base_price = ?, quantity = ?, cost_price = ?, properties = ?
		WHERE id = ?
	`

	_, err = r.db.ExecContext(ctx, query,
		product.Name,
		product.SKU,
		sql.NullString{String: product.CategoryID, Valid: product.CategoryID != ""},
		product.BasePrice,
		product.Quantity,
		product.CostPrice,
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
		ID:         row.ID,
		Name:       row.Name,
		SKU:        row.SKU,
		CategoryID: row.CategoryID.String,
		BasePrice:  row.BasePrice,
		Quantity:   row.Quantity,
		CostPrice:  row.CostPrice,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
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
