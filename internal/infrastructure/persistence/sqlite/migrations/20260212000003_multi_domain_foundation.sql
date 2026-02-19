-- +goose Up
-- Migration: 004_multi_domain_foundation
-- Description: Add multi-domain infrastructure while preserving all existing finance data
-- SAFETY: Only adds new tables, never modifies existing ones

-- Universal entity tracking table
-- This creates a unified view across all life domains
CREATE TABLE IF NOT EXISTS entities (
    id TEXT PRIMARY KEY,
    domain TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_subtype TEXT,
    title TEXT NOT NULL,
    status TEXT DEFAULT 'active',
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP,
    metadata TEXT -- JSON blob for domain-specific data
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_entities_domain ON entities(domain);
CREATE INDEX IF NOT EXISTS idx_entities_type ON entities(domain, entity_type);
CREATE INDEX IF NOT EXISTS idx_entities_status ON entities(status);
CREATE INDEX IF NOT EXISTS idx_entities_created ON entities(created_at);
CREATE INDEX IF NOT EXISTS idx_entities_updated ON entities(updated_at);

-- Cross-domain relationships
-- Enables linking between different life domains
CREATE TABLE IF NOT EXISTS entity_relationships (
    id TEXT PRIMARY KEY,
    from_entity_id TEXT NOT NULL,
    to_entity_id TEXT NOT NULL,
    relationship_type TEXT NOT NULL,
    strength INTEGER DEFAULT 1 CHECK(strength >= 1 AND strength <= 5),
    bidirectional INTEGER DEFAULT 0,
    metadata TEXT, -- JSON for relationship-specific data
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (from_entity_id) REFERENCES entities(id) ON DELETE CASCADE,
    FOREIGN KEY (to_entity_id) REFERENCES entities(id) ON DELETE CASCADE,

    -- Ensure no self-references
    CHECK (from_entity_id != to_entity_id),
    UNIQUE (from_entity_id, to_entity_id, relationship_type)
);

-- Indexes for relationships
CREATE INDEX IF NOT EXISTS idx_relationships_from ON entity_relationships(from_entity_id);
CREATE INDEX IF NOT EXISTS idx_relationships_to ON entity_relationships(to_entity_id);
CREATE INDEX IF NOT EXISTS idx_relationships_type ON entity_relationships(relationship_type);

-- Universal tag system
-- Supports both domain-specific and global tags
CREATE TABLE IF NOT EXISTS tags (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    domain TEXT, -- NULL = global tag, specific value = domain-specific
    color TEXT DEFAULT '#666666',
    description TEXT,
    usage_count INTEGER DEFAULT 0,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(name, domain)
);

CREATE TABLE IF NOT EXISTS entity_tags (
    entity_id TEXT NOT NULL,
    tag_id TEXT NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (entity_id, tag_id),
    FOREIGN KEY (entity_id) REFERENCES entities(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- Indexes for tags
CREATE INDEX IF NOT EXISTS idx_tags_domain ON tags(domain);
CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
CREATE INDEX IF NOT EXISTS idx_tags_usage ON tags(usage_count);
CREATE INDEX IF NOT EXISTS idx_entity_tags_entity ON entity_tags(entity_id);
CREATE INDEX IF NOT EXISTS idx_entity_tags_tag ON entity_tags(tag_id);

-- Simple search index table (fallback without FTS5)
-- Enables searching across all domains
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

-- Indexes for search performance
CREATE INDEX IF NOT EXISTS idx_search_domain ON search_index(domain);
CREATE INDEX IF NOT EXISTS idx_search_type ON search_index(entity_type);
CREATE INDEX IF NOT EXISTS idx_search_title ON search_index(title);

-- Activity log for tracking changes across domains
CREATE TABLE IF NOT EXISTS activity_log (
    id TEXT PRIMARY KEY,
    entity_id TEXT,
    action TEXT NOT NULL, -- 'create', 'update', 'delete', 'tag', 'link'
    domain TEXT NOT NULL,
    description TEXT,
    metadata TEXT, -- JSON for action-specific data
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (entity_id) REFERENCES entities(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_activity_entity ON activity_log(entity_id);
CREATE INDEX IF NOT EXISTS idx_activity_domain ON activity_log(domain);
CREATE INDEX IF NOT EXISTS idx_activity_created ON activity_log(created_at);
CREATE INDEX IF NOT EXISTS idx_activity_action ON activity_log(action);

-- POPULATE with existing finance data
-- This creates entity records that REFERENCE existing data without duplicating it

-- Map existing financial accounts to entities
INSERT OR IGNORE INTO entities (id, domain, entity_type, entity_subtype, title, created_at, metadata)
SELECT
    'finance_account_' || id,
    'finance',
    'account',
    LOWER(COALESCE(type, 'unknown')),
    name,
    CURRENT_TIMESTAMP,
    json_object(
        'original_id', id,
        'account_type', type,
        'credit_limit', credit_limit,
        'active', active
    )
FROM financial_accounts
WHERE name IS NOT NULL AND name != '';

-- Map recent transactions to entities (limit for performance)
INSERT OR IGNORE INTO entities (id, domain, entity_type, entity_subtype, title, created_at, metadata)
SELECT
    'finance_transaction_' || id,
    'finance',
    'transaction',
    LOWER(type),
    CASE
        WHEN name != '' AND name IS NOT NULL THEN name
        ELSE account || ' - $' || printf('%.2f', amount)
    END,
    COALESCE(created_at, CURRENT_TIMESTAMP),
    json_object(
        'original_id', id,
        'amount', amount,
        'account', account,
        'category', category,
        'transaction_type', type,
        'date', date
    )
FROM finance_transactions
WHERE date >= date('now', '-2 years') -- Recent transactions only for performance
ORDER BY date DESC;

-- Populate search index with finance data
INSERT OR REPLACE INTO search_index (entity_id, domain, entity_type, title, content, tags)
SELECT
    'finance_account_' || a.id,
    'finance',
    'account',
    a.name,
    a.name || ' ' || COALESCE(a.type, '') || ' account',
    COALESCE(a.type, '')
FROM financial_accounts a
WHERE a.name IS NOT NULL AND a.name != '';

INSERT OR REPLACE INTO search_index (entity_id, domain, entity_type, title, content, tags)
SELECT
    'finance_transaction_' || t.id,
    'finance',
    'transaction',
    CASE
        WHEN t.name != '' AND t.name IS NOT NULL THEN t.name
        ELSE t.account || ' transaction'
    END,
    COALESCE(t.name, '') || ' ' ||
    COALESCE(t.account, '') || ' ' ||
    COALESCE(t.category, '') || ' ' ||
    COALESCE(t.parent_category, '') || ' ' ||
    COALESCE(t.note, '') || ' ' ||
    '$' || printf('%.2f', t.amount),
    COALESCE(t.tags, '') || ' ' || COALESCE(t.category, '') || ' ' || COALESCE(t.type, '')
FROM finance_transactions t
WHERE t.date >= date('now', '-2 years')
ORDER BY t.date DESC;

-- Create useful default tags
INSERT OR IGNORE INTO tags (id, name, domain, color, description) VALUES
    ('tag_finance_income', 'income', 'finance', '#4CAF50', 'Income transactions'),
    ('tag_finance_expense', 'expense', 'finance', '#F44336', 'Expense transactions'),
    ('tag_finance_investment', 'investment', 'finance', '#2196F3', 'Investment related'),
    ('tag_finance_recurring', 'recurring', 'finance', '#FF9800', 'Recurring payments'),
    ('tag_finance_important', 'important', 'finance', '#E91E63', 'Important financial items'),
    ('tag_global_urgent', 'urgent', NULL, '#FF5722', 'Urgent items requiring attention'),
    ('tag_global_important', 'important', NULL, '#E91E63', 'Important items across all domains'),
    ('tag_global_archived', 'archived', NULL, '#9E9E9E', 'Archived items'),
    ('tag_global_favorite', 'favorite', NULL, '#FFC107', 'Favorite items'),
    ('tag_global_todo', 'todo', NULL, '#03A9F4', 'Items that need action');

-- Log the migration completion
INSERT INTO activity_log (id, action, domain, description, metadata) VALUES (
    'activity_' || hex(randomblob(8)),
    'migrate',
    'system',
    'Multi-domain infrastructure migration completed',
    json_object(
        'migration', '004_multi_domain_foundation',
        'entities_created', (SELECT COUNT(*) FROM entities),
        'tags_created', (SELECT COUNT(*) FROM tags),
        'finance_accounts_mapped', (SELECT COUNT(*) FROM entities WHERE domain = 'finance' AND entity_type = 'account'),
        'finance_transactions_mapped', (SELECT COUNT(*) FROM entities WHERE domain = 'finance' AND entity_type = 'transaction')
    )
);

-- +goose Down
-- No down migration for multi-domain foundation - data would need manual cleanup
