package ports

import (
	"context"

	"github.com/torantous1337/retail-management/internal/core/domain"
)

// ProductService defines the interface for product business logic.
type ProductService interface {
	CreateProduct(ctx context.Context, product *domain.Product) error
	GetProduct(ctx context.Context, id string) (*domain.Product, error)
	GetProductBySKU(ctx context.Context, sku string) (*domain.Product, error)
	ListProducts(ctx context.Context, limit, offset int) ([]*domain.Product, error)
	UpdateProduct(ctx context.Context, product *domain.Product) error
	DeleteProduct(ctx context.Context, id string) error
}

// AuditService defines the interface for audit logging with tamper-proofing.
type AuditService interface {
	LogAction(ctx context.Context, action, userID string, payload map[string]interface{}) error
	VerifyAuditChain(ctx context.Context) (bool, error)
	GetAuditLogs(ctx context.Context, limit, offset int) ([]*domain.AuditLog, error)
}
