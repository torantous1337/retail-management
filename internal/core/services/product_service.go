package services

import (
	"context"

	"github.com/torantous1337/retail-management/internal/core/domain"
	"github.com/torantous1337/retail-management/internal/core/ports"
)

// ProductService implements the product business logic.
type ProductService struct {
	productRepo ports.ProductRepository
	auditSvc    ports.AuditService
}

// NewProductService creates a new product service instance.
func NewProductService(productRepo ports.ProductRepository, auditSvc ports.AuditService) *ProductService {
	return &ProductService{
		productRepo: productRepo,
		auditSvc:    auditSvc,
	}
}

// CreateProduct creates a new product and logs the action.
func (s *ProductService) CreateProduct(ctx context.Context, product *domain.Product) error {
	err := s.productRepo.Create(ctx, product)
	if err != nil {
		return err
	}

	// Log the action (audit failures should not prevent product creation but should be logged)
	payload := map[string]interface{}{
		"product_id": product.ID,
		"action":     "create_product",
		"sku":        product.SKU,
		"name":       product.Name,
	}
	if err := s.auditSvc.LogAction(ctx, "CREATE_PRODUCT", "system", payload); err != nil {
		// TODO: Log this error to a monitoring system
		// For now, we don't fail the operation but the error should be tracked
	}

	return nil
}

// GetProduct retrieves a product by ID.
func (s *ProductService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	return s.productRepo.GetByID(ctx, id)
}

// GetProductBySKU retrieves a product by SKU.
func (s *ProductService) GetProductBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	return s.productRepo.GetBySKU(ctx, sku)
}

// ListProducts retrieves all products with pagination.
func (s *ProductService) ListProducts(ctx context.Context, limit, offset int) ([]*domain.Product, error) {
	return s.productRepo.List(ctx, limit, offset)
}

// UpdateProduct updates a product and logs the action.
func (s *ProductService) UpdateProduct(ctx context.Context, product *domain.Product) error {
	err := s.productRepo.Update(ctx, product)
	if err != nil {
		return err
	}

	// Log the action (audit failures should not prevent product update but should be logged)
	payload := map[string]interface{}{
		"product_id": product.ID,
		"action":     "update_product",
		"sku":        product.SKU,
	}
	if err := s.auditSvc.LogAction(ctx, "UPDATE_PRODUCT", "system", payload); err != nil {
		// TODO: Log this error to a monitoring system
		// For now, we don't fail the operation but the error should be tracked
	}

	return nil
}

// DeleteProduct deletes a product and logs the action.
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	err := s.productRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Log the action (audit failures should not prevent product deletion but should be logged)
	payload := map[string]interface{}{
		"product_id": id,
		"action":     "delete_product",
	}
	if err := s.auditSvc.LogAction(ctx, "DELETE_PRODUCT", "system", payload); err != nil {
		// TODO: Log this error to a monitoring system
		// For now, we don't fail the operation but the error should be tracked
	}

	return nil
}
