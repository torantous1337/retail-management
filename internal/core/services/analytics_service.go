package services

import (
	"context"

	"github.com/torantous1337/retail-management/internal/core/domain"
	"github.com/torantous1337/retail-management/internal/core/ports"
)

// AnalyticsService implements the analytics business logic.
type AnalyticsService struct {
	productRepo ports.ProductRepository
}

// NewAnalyticsService creates a new analytics service instance.
func NewAnalyticsService(productRepo ports.ProductRepository) *AnalyticsService {
	return &AnalyticsService{
		productRepo: productRepo,
	}
}

// GetInventorySummary returns aggregated inventory analytics.
func (s *AnalyticsService) GetInventorySummary(ctx context.Context) (*domain.InventorySummary, error) {
	return s.productRepo.GetInventorySummary(ctx)
}
