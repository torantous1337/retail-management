-- Migration 003: Full-Text Search (FTS5)
-- Adds full-text search support on product name and sku.

CREATE VIRTUAL TABLE IF NOT EXISTS products_fts USING fts5(
    name,
    sku,
    content='products',
    content_rowid='rowid'
);

-- Triggers to keep the FTS index in sync with the products table.

CREATE TRIGGER IF NOT EXISTS products_ai AFTER INSERT ON products BEGIN
    INSERT INTO products_fts(rowid, name, sku) VALUES (new.rowid, new.name, new.sku);
END;

CREATE TRIGGER IF NOT EXISTS products_ad AFTER DELETE ON products BEGIN
    INSERT INTO products_fts(products_fts, rowid, name, sku) VALUES('delete', old.rowid, old.name, old.sku);
END;

CREATE TRIGGER IF NOT EXISTS products_au AFTER UPDATE ON products BEGIN
    INSERT INTO products_fts(products_fts, rowid, name, sku) VALUES('delete', old.rowid, old.name, old.sku);
    INSERT INTO products_fts(rowid, name, sku) VALUES (new.rowid, new.name, new.sku);
END;
