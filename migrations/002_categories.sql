-- Migration 002: Categories (Dynamic Category Blueprint)
-- Adds category support for enforcing product property schemas.

-- Categories table
CREATE TABLE IF NOT EXISTS categories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    attribute_definitions TEXT -- JSON array of attribute definitions
);

-- Add category_id to products with foreign key reference
ALTER TABLE products ADD COLUMN category_id TEXT REFERENCES categories(id);
