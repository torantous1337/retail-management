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
