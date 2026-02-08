package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/torantous1337/retail-management/internal/core/domain"
)

// CategoryRepository implements the category repository using SQLite.
type CategoryRepository struct {
	db *sqlx.DB
}

// NewCategoryRepository creates a new category repository instance.
func NewCategoryRepository(db *sqlx.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// categoryRow is a database row representation for categories.
type categoryRow struct {
	ID                   string         `db:"id"`
	Name                 string         `db:"name"`
	AttributeDefinitions sql.NullString `db:"attribute_definitions"`
}

// Create creates a new category in the database.
func (r *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	attrsJSON, err := json.Marshal(category.AttributeDefinitions)
	if err != nil {
		return err
	}

	query := `INSERT INTO categories (id, name, attribute_definitions) VALUES (?, ?, ?)`
	_, err = r.db.ExecContext(ctx, query, category.ID, category.Name, string(attrsJSON))
	return err
}

// GetByID retrieves a category by its ID.
func (r *CategoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	query := `SELECT * FROM categories WHERE id = ?`

	var row categoryRow
	err := r.db.GetContext(ctx, &row, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}

	return r.toDomain(&row)
}

// List retrieves all categories with pagination.
func (r *CategoryRepository) List(ctx context.Context, limit, offset int) ([]*domain.Category, error) {
	query := `SELECT * FROM categories ORDER BY name LIMIT ? OFFSET ?`

	var rows []categoryRow
	err := r.db.SelectContext(ctx, &rows, query, limit, offset)
	if err != nil {
		return nil, err
	}

	categories := make([]*domain.Category, 0, len(rows))
	for _, row := range rows {
		category, err := r.toDomain(&row)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// toDomain converts a database row to a domain entity.
func (r *CategoryRepository) toDomain(row *categoryRow) (*domain.Category, error) {
	category := &domain.Category{
		ID:   row.ID,
		Name: row.Name,
	}

	if row.AttributeDefinitions.Valid && row.AttributeDefinitions.String != "" {
		var attrs []domain.AttributeDefinition
		err := json.Unmarshal([]byte(row.AttributeDefinitions.String), &attrs)
		if err != nil {
			return nil, err
		}
		category.AttributeDefinitions = attrs
	} else {
		category.AttributeDefinitions = []domain.AttributeDefinition{}
	}

	return category, nil
}
