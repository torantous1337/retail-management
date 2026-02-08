-- Migration 004: Sales and Inventory
-- Adds inventory tracking (quantity, cost_price) to products and creates sales tables.

-- Add inventory fields to products
ALTER TABLE products ADD COLUMN quantity INTEGER NOT NULL DEFAULT 0;
ALTER TABLE products ADD COLUMN cost_price REAL NOT NULL DEFAULT 0.0;

-- Sales table
CREATE TABLE IF NOT EXISTS sales (
    id TEXT PRIMARY KEY,
    total_amount REAL NOT NULL DEFAULT 0.0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for reporting queries on sales
CREATE INDEX IF NOT EXISTS idx_sales_created_at ON sales(created_at);

-- Sale items table
CREATE TABLE IF NOT EXISTS sale_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sale_id TEXT NOT NULL REFERENCES sales(id),
    product_id TEXT NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    unit_price REAL NOT NULL,
    cost_price REAL NOT NULL
);
