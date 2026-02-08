package ports

import (
	"context"

	"github.com/torantous1337/retail-management/internal/core/domain"
)

// ProductRepository defines the interface for product data access.
type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	GetBySKU(ctx context.Context, sku string) (*domain.Product, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Product, error)
	Search(ctx context.Context, opts domain.FilterOptions, allowedKeys []string) ([]*domain.Product, error)
	GetInventorySummary(ctx context.Context) (*domain.InventorySummary, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, id string) error
}

// CategoryRepository defines the interface for category data access.
type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, id string) (*domain.Category, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Category, error)
}

// AuditLogRepository defines the interface for audit log data access.
type AuditLogRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	GetLastLog(ctx context.Context) (*domain.AuditLog, error)
	List(ctx context.Context, limit, offset int) ([]*domain.AuditLog, error)
	VerifyChain(ctx context.Context) (bool, error)
}

// Ports bundles all repository interfaces for use in transactions.
type Ports struct {
	ProductRepo  ProductRepository
	CategoryRepo CategoryRepository
	AuditRepo    AuditLogRepository
}

// TransactionManager provides atomic transaction support.
type TransactionManager interface {
	WithTx(ctx context.Context, fn func(tx Ports) error) error
}
