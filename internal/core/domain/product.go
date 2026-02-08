package domain

import "time"

// Product represents a product entity in the system.
// This is a pure business entity with no framework tags.
type Product struct {
	ID         string
	Name       string
	SKU        string
	CategoryID string
	BasePrice  float64
	Properties map[string]interface{} // Flexible attributes (voltage, amperage, etc.)
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// FilterOptions holds the parameters for searching and filtering products.
type FilterOptions struct {
	Query      string            // Full-text search query on name/sku
	CategoryID string            // Filter by category
	MinPrice   *float64          // Minimum base_price
	MaxPrice   *float64          // Maximum base_price
	Properties map[string]string // Dynamic JSON property filters (key -> value)
	Limit      int
	Offset     int
}

// CategoryBreakdown holds aggregated data for a single category.
type CategoryBreakdown struct {
	CategoryID   string
	CategoryName string
	Count        int
	TotalValue   float64
}

// InventorySummary holds aggregated inventory analytics.
type InventorySummary struct {
	TotalItems      int
	TotalValue      float64
	CategoryBreakdown []CategoryBreakdown
}
