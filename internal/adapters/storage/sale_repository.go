package storage

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/torantous1337/retail-management/internal/core/domain"
)

// SaleRepository implements the sale repository using SQLite.
type SaleRepository struct {
	db sqlx.ExtContext
}

// NewSaleRepository creates a new sale repository instance.
func NewSaleRepository(db sqlx.ExtContext) *SaleRepository {
	return &SaleRepository{db: db}
}

// CreateSale inserts a new sale record.
func (r *SaleRepository) CreateSale(ctx context.Context, sale *domain.Sale) error {
	query := `INSERT INTO sales (id, total_amount, created_at) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, sale.ID, sale.TotalAmount, sale.CreatedAt)
	return err
}

// CreateSaleItem inserts a new sale item record.
func (r *SaleRepository) CreateSaleItem(ctx context.Context, item *domain.SaleItem) error {
	query := `INSERT INTO sale_items (sale_id, product_id, quantity, unit_price, cost_price) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, item.SaleID, item.ProductID, item.Quantity, item.UnitPrice, item.CostPrice)
	return err
}
