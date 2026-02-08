package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/torantous1337/retail-management/internal/core/domain"
	"github.com/torantous1337/retail-management/internal/core/ports"
)

// ErrInvalidProperty is returned when product properties fail category validation.
var ErrInvalidProperty = errors.New("invalid property")

// CategoryService implements the category business logic.
type CategoryService struct {
	categoryRepo ports.CategoryRepository
}

// NewCategoryService creates a new category service instance.
func NewCategoryService(categoryRepo ports.CategoryRepository) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

// CreateCategory creates a new category.
func (s *CategoryService) CreateCategory(ctx context.Context, category *domain.Category) error {
	return s.categoryRepo.Create(ctx, category)
}

// GetCategory retrieves a category by ID.
func (s *CategoryService) GetCategory(ctx context.Context, id string) (*domain.Category, error) {
	return s.categoryRepo.GetByID(ctx, id)
}

// ListCategories retrieves all categories with pagination.
func (s *CategoryService) ListCategories(ctx context.Context, limit, offset int) ([]*domain.Category, error) {
	return s.categoryRepo.List(ctx, limit, offset)
}

// ValidateProperties validates product properties against the category's attribute definitions.
func ValidateProperties(category *domain.Category, properties map[string]interface{}) error {
	if category == nil {
		return nil
	}

	for _, attr := range category.AttributeDefinitions {
		val, exists := properties[attr.Key]

		// Check required fields
		if attr.Required && !exists {
			return fmt.Errorf("%w: missing required property %q", ErrInvalidProperty, attr.Key)
		}

		if !exists {
			continue
		}

		// Check data types
		switch attr.Type {
		case "string":
			if _, ok := val.(string); !ok {
				return fmt.Errorf("%w: property %q must be a string", ErrInvalidProperty, attr.Key)
			}
		case "number":
			if !isNumeric(val) {
				return fmt.Errorf("%w: property %q must be a number", ErrInvalidProperty, attr.Key)
			}
		case "boolean":
			if _, ok := val.(bool); !ok {
				return fmt.Errorf("%w: property %q must be a boolean", ErrInvalidProperty, attr.Key)
			}
		case "select":
			strVal, ok := val.(string)
			if !ok {
				return fmt.Errorf("%w: property %q must be a string for select type", ErrInvalidProperty, attr.Key)
			}
			if len(attr.Options) > 0 && !contains(attr.Options, strVal) {
				return fmt.Errorf("%w: property %q value %q is not in allowed options %v", ErrInvalidProperty, attr.Key, strVal, attr.Options)
			}
		}
	}

	return nil
}

// isNumeric checks if a value is a numeric type.
func isNumeric(val interface{}) bool {
	switch v := val.(type) {
	case float64, float32, int, int32, int64:
		return true
	case string:
		_, err := strconv.ParseFloat(v, 64)
		return err == nil
	default:
		return false
	}
}

// contains checks if a string is in a slice.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
