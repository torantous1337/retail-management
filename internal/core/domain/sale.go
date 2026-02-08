package domain

import "time"

// Sale represents a completed sales transaction.
type Sale struct {
	ID          string
	TotalAmount float64
	CreatedAt   time.Time
}

// SaleItem represents a single line item in a sale.
type SaleItem struct {
	SaleID    string
	ProductID string
	Quantity  int
	UnitPrice float64 // Snapshot of BasePrice at time of sale
	CostPrice float64 // Snapshot of CostPrice at time of sale
}
