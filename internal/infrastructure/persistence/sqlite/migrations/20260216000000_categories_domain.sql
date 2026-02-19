-- +goose Up
-- Categories table with hierarchy support
-- Date: 2026-02-16

CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    parent_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    domain TEXT NOT NULL DEFAULT 'finance',
    description TEXT,
    is_active INTEGER DEFAULT 1,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    UNIQUE(name, domain)
);

CREATE INDEX IF NOT EXISTS idx_categories_domain ON categories(domain);
CREATE INDEX IF NOT EXISTS idx_categories_parent ON categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_categories_name_domain ON categories(name, domain);

-- Insert default Uncategorized category for finance domain if not exists
INSERT OR IGNORE INTO categories (name, domain, description, is_active)
VALUES ('Uncategorized', 'finance', 'Default category for uncategorized items', 1);

-- +goose Down
-- DROP TABLE IF EXISTS categories;
