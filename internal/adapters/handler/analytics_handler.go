package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/torantous1337/retail-management/internal/core/ports"
)

// AnalyticsHandler handles HTTP requests for analytics.
type AnalyticsHandler struct {
	analyticsSvc ports.AnalyticsService
}

// NewAnalyticsHandler creates a new analytics handler instance.
func NewAnalyticsHandler(analyticsSvc ports.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsSvc: analyticsSvc,
	}
}

// GetInventorySummary handles GET /analytics/summary
func (h *AnalyticsHandler) GetInventorySummary(c *fiber.Ctx) error {
	summary, err := h.analyticsSvc.GetInventorySummary(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get inventory summary",
		})
	}

	breakdown := make([]fiber.Map, 0, len(summary.CategoryBreakdown))
	for _, cb := range summary.CategoryBreakdown {
		breakdown = append(breakdown, fiber.Map{
			"category_id":   cb.CategoryID,
			"category_name": cb.CategoryName,
			"count":         cb.Count,
			"total_value":   cb.TotalValue,
		})
	}

	return c.JSON(fiber.Map{
		"total_items":        summary.TotalItems,
		"total_value":        summary.TotalValue,
		"category_breakdown": breakdown,
	})
}
