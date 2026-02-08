package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/torantous1337/retail-management/internal/core/domain"
	"github.com/torantous1337/retail-management/internal/core/ports"
)

// ErrInsufficientStock is returned when a product has insufficient stock for a sale.
var ErrInsufficientStock = errors.New("insufficient stock")

// SaleService implements the sale processing logic.
type SaleService struct {
	txManager ports.TransactionManager
}

// NewSaleService creates a new sale service instance.
func NewSaleService(txManager ports.TransactionManager) *SaleService {
	return &SaleService{
		txManager: txManager,
	}
}

// ProcessSale executes an atomic checkout: validates stock, decrements quantities,
// creates sale items with price snapshots, and records the sale with an audit log.
func (s *SaleService) ProcessSale(ctx context.Context, items []ports.SaleItemRequest) (*domain.Sale, error) {
	if len(items) == 0 {
		return nil, errors.New("no items in sale")
	}

	sale := &domain.Sale{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
	}

	err := s.txManager.WithTx(ctx, func(tx ports.Ports) error {
		var total float64

		for _, item := range items {
			if item.Quantity <= 0 {
				return fmt.Errorf("invalid quantity for product %s", item.ProductID)
			}

			product, err := tx.ProductRepo.GetByID(ctx, item.ProductID)
			if err != nil {
				return fmt.Errorf("product %s: %w", item.ProductID, err)
			}

			if product.Quantity < item.Quantity {
				return fmt.Errorf("%w: product %s has %d in stock, requested %d",
					ErrInsufficientStock, item.ProductID, product.Quantity, item.Quantity)
			}

			// Decrement stock
			product.Quantity -= item.Quantity
			if err := tx.ProductRepo.Update(ctx, product); err != nil {
				return fmt.Errorf("update stock for product %s: %w", item.ProductID, err)
			}

			// Create sale item with price snapshots
			saleItem := &domain.SaleItem{
				SaleID:    sale.ID,
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				UnitPrice: product.BasePrice,
				CostPrice: product.CostPrice,
			}
			if err := tx.SaleRepo.CreateSaleItem(ctx, saleItem); err != nil {
				return fmt.Errorf("create sale item for product %s: %w", item.ProductID, err)
			}

			total += product.BasePrice * float64(item.Quantity)
		}

		sale.TotalAmount = total
		if err := tx.SaleRepo.CreateSale(ctx, sale); err != nil {
			return fmt.Errorf("create sale: %w", err)
		}

		// Audit log
		lastLog, err := tx.AuditRepo.GetLastLog(ctx)
		prevHash := ""
		if err == nil && lastLog != nil {
			prevHash = lastLog.CurrentHash
		}

		txAuditSvc := NewAuditService(tx.AuditRepo)
		txAuditSvc.SetPrevHash(prevHash)
		if err := txAuditSvc.LogAction(ctx, "SALE_PROCESSED", "system", map[string]interface{}{
			"sale_id":      sale.ID,
			"total_amount": sale.TotalAmount,
			"item_count":   len(items),
		}); err != nil {
			return fmt.Errorf("audit log: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return sale, nil
}
