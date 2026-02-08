package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/torantous1337/retail-management/internal/core/ports"
	"github.com/torantous1337/retail-management/internal/core/services"
)

// SaleHandler handles HTTP requests for sales.
type SaleHandler struct {
	saleSvc ports.SaleService
}

// NewSaleHandler creates a new sale handler instance.
func NewSaleHandler(saleSvc ports.SaleService) *SaleHandler {
	return &SaleHandler{
		saleSvc: saleSvc,
	}
}

// processSaleRequest represents the request body for processing a sale.
type processSaleRequest struct {
	Items []processSaleItemRequest `json:"items"`
}

// processSaleItemRequest represents a single item in a sale request.
type processSaleItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// saleResponse represents the response body for a sale.
type saleResponse struct {
	ID          string  `json:"id"`
	TotalAmount float64 `json:"total_amount"`
	CreatedAt   time.Time `json:"created_at"`
}

// ProcessSale handles POST /api/v1/sales
func (h *SaleHandler) ProcessSale(c *fiber.Ctx) error {
	var req processSaleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one item is required",
		})
	}

	// Convert to service request
	saleItems := make([]ports.SaleItemRequest, len(req.Items))
	for i, item := range req.Items {
		if item.ProductID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "product_id is required for each item",
			})
		}
		if item.Quantity <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "quantity must be positive for each item",
			})
		}
		saleItems[i] = ports.SaleItemRequest{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	sale, err := h.saleSvc.ProcessSale(c.Context(), saleItems)
	if err != nil {
		if errors.Is(err, services.ErrInsufficientStock) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(saleResponse{
		ID:          sale.ID,
		TotalAmount: sale.TotalAmount,
		CreatedAt:   sale.CreatedAt,
	})
}
