package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/torantous1337/retail-management/internal/core/domain"
	"github.com/torantous1337/retail-management/internal/core/ports"
	"github.com/torantous1337/retail-management/internal/core/services"
)

// ProductHandler handles HTTP requests for products.
type ProductHandler struct {
	productSvc ports.ProductService
}

// NewProductHandler creates a new product handler instance.
func NewProductHandler(productSvc ports.ProductService) *ProductHandler {
	return &ProductHandler{
		productSvc: productSvc,
	}
}

// CreateProductRequest represents the request body for creating a product.
type CreateProductRequest struct {
	Name       string                 `json:"name"`
	SKU        string                 `json:"sku"`
	CategoryID string                 `json:"category_id"`
	BasePrice  float64                `json:"base_price"`
	Properties map[string]interface{} `json:"properties"`
}

// ProductResponse represents the response body for a product.
type ProductResponse struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	SKU        string                 `json:"sku"`
	CategoryID string                 `json:"category_id,omitempty"`
	BasePrice  float64                `json:"base_price"`
	Properties map[string]interface{} `json:"properties"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// CreateProduct handles POST /products
func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var req CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Name == "" || req.SKU == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name and SKU are required",
		})
	}

	now := time.Now()
	product := &domain.Product{
		ID:         uuid.New().String(),
		Name:       req.Name,
		SKU:        req.SKU,
		CategoryID: req.CategoryID,
		BasePrice:  req.BasePrice,
		Properties: req.Properties,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	err := h.productSvc.CreateProduct(c.Context(), product)
	if err != nil {
		if errors.Is(err, services.ErrInvalidProperty) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create product",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(h.toResponse(product))
}

// GetProduct handles GET /products/:id
func (h *ProductHandler) GetProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Product ID is required",
		})
	}

	product, err := h.productSvc.GetProduct(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Product not found",
		})
	}

	return c.JSON(h.toResponse(product))
}

// GetProductBySKU handles GET /products/sku/:sku
func (h *ProductHandler) GetProductBySKU(c *fiber.Ctx) error {
	sku := c.Params("sku")
	if sku == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "SKU is required",
		})
	}

	product, err := h.productSvc.GetProductBySKU(c.Context(), sku)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Product not found",
		})
	}

	return c.JSON(h.toResponse(product))
}

// ListProducts handles GET /products
func (h *ProductHandler) ListProducts(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)

	products, err := h.productSvc.ListProducts(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list products",
		})
	}

	responses := make([]ProductResponse, 0, len(products))
	for _, product := range products {
		responses = append(responses, h.toResponse(product))
	}

	return c.JSON(fiber.Map{
		"products": responses,
		"limit":    limit,
		"offset":   offset,
	})
}

// UpdateProduct handles PUT /products/:id
func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Product ID is required",
		})
	}

	var req CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get existing product to preserve created_at
	existing, err := h.productSvc.GetProduct(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Product not found",
		})
	}

	product := &domain.Product{
		ID:         id,
		Name:       req.Name,
		SKU:        req.SKU,
		CategoryID: req.CategoryID,
		BasePrice:  req.BasePrice,
		Properties: req.Properties,
		CreatedAt:  existing.CreatedAt,
		UpdatedAt:  time.Now(),
	}

	err = h.productSvc.UpdateProduct(c.Context(), product)
	if err != nil {
		if errors.Is(err, services.ErrInvalidProperty) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update product",
		})
	}

	return c.JSON(h.toResponse(product))
}

// DeleteProduct handles DELETE /products/:id
func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Product ID is required",
		})
	}

	err := h.productSvc.DeleteProduct(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete product",
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ImportProducts handles POST /products/import
func (h *ProductHandler) ImportProducts(c *fiber.Ctx) error {
	categoryID := c.FormValue("category_id")

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File field 'file' is required",
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to open uploaded file",
		})
	}
	defer file.Close()

	count, err := h.productSvc.ImportProducts(c.Context(), categoryID, file)
	if err != nil {
		if errors.Is(err, services.ErrInvalidProperty) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"imported": count,
	})
}

// toResponse converts a domain product to a response DTO.
func (h *ProductHandler) toResponse(product *domain.Product) ProductResponse {
	return ProductResponse{
		ID:         product.ID,
		Name:       product.Name,
		SKU:        product.SKU,
		CategoryID: product.CategoryID,
		BasePrice:  product.BasePrice,
		Properties: product.Properties,
		CreatedAt:  product.CreatedAt,
		UpdatedAt:  product.UpdatedAt,
	}
}
