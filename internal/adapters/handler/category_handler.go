package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/torantous1337/retail-management/internal/core/domain"
	"github.com/torantous1337/retail-management/internal/core/ports"
)

// CategoryHandler handles HTTP requests for categories.
type CategoryHandler struct {
	categorySvc ports.CategoryService
}

// NewCategoryHandler creates a new category handler instance.
func NewCategoryHandler(categorySvc ports.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categorySvc: categorySvc,
	}
}

// AttributeDefinitionRequest represents an attribute definition in a request.
type AttributeDefinitionRequest struct {
	Key      string   `json:"key"`
	Type     string   `json:"type"`
	Required bool     `json:"required"`
	Options  []string `json:"options"`
	Unit     string   `json:"unit"`
}

// CreateCategoryRequest represents the request body for creating a category.
type CreateCategoryRequest struct {
	Name                 string                       `json:"name"`
	AttributeDefinitions []AttributeDefinitionRequest `json:"attribute_definitions"`
}

// CategoryResponse represents the response body for a category.
type CategoryResponse struct {
	ID                   string                       `json:"id"`
	Name                 string                       `json:"name"`
	AttributeDefinitions []AttributeDefinitionRequest `json:"attribute_definitions"`
}

// CreateCategory handles POST /categories
func (h *CategoryHandler) CreateCategory(c *fiber.Ctx) error {
	var req CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}

	attrs := make([]domain.AttributeDefinition, 0, len(req.AttributeDefinitions))
	for _, a := range req.AttributeDefinitions {
		attrs = append(attrs, domain.AttributeDefinition{
			Key:      a.Key,
			Type:     a.Type,
			Required: a.Required,
			Options:  a.Options,
			Unit:     a.Unit,
		})
	}

	category := &domain.Category{
		ID:                   uuid.New().String(),
		Name:                 req.Name,
		AttributeDefinitions: attrs,
	}

	err := h.categorySvc.CreateCategory(c.Context(), category)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create category",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(h.toResponse(category))
}

// ListCategories handles GET /categories
func (h *CategoryHandler) ListCategories(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)

	categories, err := h.categorySvc.ListCategories(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list categories",
		})
	}

	responses := make([]CategoryResponse, 0, len(categories))
	for _, category := range categories {
		responses = append(responses, h.toResponse(category))
	}

	return c.JSON(fiber.Map{
		"categories": responses,
		"limit":      limit,
		"offset":     offset,
	})
}

// toResponse converts a domain category to a response DTO.
func (h *CategoryHandler) toResponse(category *domain.Category) CategoryResponse {
	attrs := make([]AttributeDefinitionRequest, 0, len(category.AttributeDefinitions))
	for _, a := range category.AttributeDefinitions {
		attrs = append(attrs, AttributeDefinitionRequest{
			Key:      a.Key,
			Type:     a.Type,
			Required: a.Required,
			Options:  a.Options,
			Unit:     a.Unit,
		})
	}

	return CategoryResponse{
		ID:                   category.ID,
		Name:                 category.Name,
		AttributeDefinitions: attrs,
	}
}
