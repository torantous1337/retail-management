package storage

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/torantous1337/retail-management/internal/core/ports"
)

// SQLTransactionManager implements TransactionManager using sqlx.
type SQLTransactionManager struct {
	db *sqlx.DB
}

// NewSQLTransactionManager creates a new transaction manager.
func NewSQLTransactionManager(db *sqlx.DB) *SQLTransactionManager {
	return &SQLTransactionManager{db: db}
}

// WithTx executes fn within a database transaction. If fn returns an error
// or panics, the transaction is rolled back; otherwise it is committed.
func (m *SQLTransactionManager) WithTx(ctx context.Context, fn func(tx ports.Ports) error) error {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	txPorts := ports.Ports{
		ProductRepo:  NewProductRepository(tx),
		CategoryRepo: NewCategoryRepository(tx),
		AuditRepo:    NewAuditLogRepository(tx),
		SaleRepo:     NewSaleRepository(tx),
	}

	if err := fn(txPorts); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
