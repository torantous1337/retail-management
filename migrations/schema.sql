-- Retail Management System Schema
-- Database: SQLite
-- Design: Local-first with replication readiness

-- Products table with flexible schema (EAV hybrid)
CREATE TABLE IF NOT EXISTS products (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    sku TEXT UNIQUE NOT NULL,
    base_price REAL NOT NULL DEFAULT 0.0,
    properties TEXT, -- JSON column for category-specific attributes
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for SKU lookups
CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);

-- Audit logs table with blockchain-like hashing for tamper-proofing
CREATE TABLE IF NOT EXISTS audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    action TEXT NOT NULL,
    user_id TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    payload TEXT NOT NULL, -- JSON payload of the change
    prev_hash TEXT, -- Hash of previous record (NULL for first record)
    current_hash TEXT NOT NULL -- SHA256(payload + timestamp + prev_hash)
);

-- Index for chronological queries
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp);

-- Index for user-specific queries
CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user_id);

-- Trigger to update updated_at timestamp on products
CREATE TRIGGER IF NOT EXISTS update_products_timestamp 
AFTER UPDATE ON products
FOR EACH ROW
BEGIN
    UPDATE products SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
