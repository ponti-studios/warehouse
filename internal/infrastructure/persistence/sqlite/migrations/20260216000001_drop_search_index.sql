-- +goose Up
-- Migration: Drop search_index table
-- Description: Removes the global search index table as global search has been removed
-- Date: 2026-02-16

DROP TABLE IF EXISTS search_index;

-- +goose Down
-- Migration: Recreate search_index table
-- Description: Recreates the search_index table (non-FTS version)

CREATE TABLE IF NOT EXISTS search_index (
    entity_id TEXT NOT NULL,
    domain TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    tags TEXT NOT NULL DEFAULT '',
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (entity_id)
);
