-- +goose Up
-- Migration: 005_tracking_domain
-- Description: Add tracking domain tables for habits, activities, health, career tracking
-- Created: February 2026

-- Tracking entries table - main table for all tracking data
CREATE TABLE IF NOT EXISTS tracking_entries (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL, -- health, career, activity, habit, goal, etc.
    title TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '', -- original markdown content
    metadata TEXT NOT NULL DEFAULT '{}', -- JSON with frontmatter and additional data
    date TEXT NOT NULL, -- primary date for the entry (ISO 8601)
    status TEXT DEFAULT 'active',
    source_file TEXT, -- original file path if imported from markdown
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key to universal entities table
    FOREIGN KEY (id) REFERENCES entities(id) ON DELETE CASCADE
);

-- Time-series data points within tracking entries
CREATE TABLE IF NOT EXISTS tracking_data_points (
    id TEXT PRIMARY KEY,
    entry_id TEXT NOT NULL,
    date TEXT NOT NULL, -- date of the data point (ISO 8601)
    type TEXT NOT NULL, -- measurement, event, goal, medication, activity, etc.
    value TEXT NOT NULL, -- flexible value field (number, text, json)
    unit TEXT DEFAULT '', -- mg, hours, count, percentage, etc.
    metadata TEXT DEFAULT '{}', -- JSON for additional structured data
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (entry_id) REFERENCES tracking_entries(id) ON DELETE CASCADE
);

-- Tracking types/categories for organization and configuration
CREATE TABLE IF NOT EXISTS tracking_types (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT DEFAULT '',
    icon TEXT DEFAULT '📝', -- emoji icon for display
    color TEXT DEFAULT '#666666', -- hex color for UI
    schema TEXT DEFAULT '{}', -- JSON schema for validation
    is_active BOOLEAN DEFAULT TRUE,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_tracking_entries_type ON tracking_entries(type);
CREATE INDEX IF NOT EXISTS idx_tracking_entries_date ON tracking_entries(date);
CREATE INDEX IF NOT EXISTS idx_tracking_entries_status ON tracking_entries(status);
CREATE INDEX IF NOT EXISTS idx_tracking_entries_created ON tracking_entries(created_at);
CREATE INDEX IF NOT EXISTS idx_tracking_entries_source_file ON tracking_entries(source_file);

CREATE INDEX IF NOT EXISTS idx_tracking_data_points_entry ON tracking_data_points(entry_id);
CREATE INDEX IF NOT EXISTS idx_tracking_data_points_date ON tracking_data_points(date);
CREATE INDEX IF NOT EXISTS idx_tracking_data_points_type ON tracking_data_points(type);
CREATE INDEX IF NOT EXISTS idx_tracking_data_points_created ON tracking_data_points(created_at);

CREATE INDEX IF NOT EXISTS idx_tracking_types_name ON tracking_types(name);
CREATE INDEX IF NOT EXISTS idx_tracking_types_active ON tracking_types(is_active);

-- Insert default tracking types
INSERT OR IGNORE INTO tracking_types (id, name, description, icon, color, is_active) VALUES
    ('tracking_type_health', 'health', 'Medical, fitness, and wellness tracking', '🏥', '#E53E3E', TRUE),
    ('tracking_type_career', 'career', 'Job applications, interviews, and career progression', '💼', '#3182CE', TRUE),
    ('tracking_type_activity', 'activity', 'Hobbies, entertainment, and recreational activities', '🎯', '#38A169', TRUE),
    ('tracking_type_habit', 'habit', 'Daily habits and routine tracking', '📈', '#805AD5', TRUE),
    ('tracking_type_goal', 'goal', 'Personal goals and progress tracking', '🎯', '#D69E2E', TRUE),
    ('tracking_type_entertainment', 'entertainment', 'Movies, shows, books, games tracking', '🎬', '#E53E3E', TRUE),
    ('tracking_type_fitness', 'fitness', 'Exercise, workouts, and physical activities', '💪', '#38A169', TRUE),
    ('tracking_type_medication', 'medication', 'Medication schedules and dosage tracking', '💊', '#4299E1', TRUE),
    ('tracking_type_therapy', 'therapy', 'Mental health and therapy sessions', '🧠', '#805AD5', TRUE),
    ('tracking_type_substance', 'substance', 'Substance use monitoring and tracking', '🚬', '#ECC94B', TRUE);

-- Create triggers to keep entities table in sync
CREATE TRIGGER IF NOT EXISTS tracking_entries_insert AFTER INSERT ON tracking_entries
BEGIN
    INSERT OR REPLACE INTO entities (id, domain, entity_type, entity_subtype, title, status, created_at, updated_at, metadata)
    VALUES (
        NEW.id,
        'tracking',
        'entry',
        NEW.type,
        NEW.title,
        NEW.status,
        NEW.created_at,
        NEW.updated_at,
        json_object(
            'source_file', NEW.source_file,
            'date', NEW.date,
            'type', NEW.type
        )
    );

    INSERT OR REPLACE INTO search_index (entity_id, domain, entity_type, title, content, tags)
    VALUES (
        NEW.id,
        'tracking',
        NEW.type,
        NEW.title,
        NEW.content,
        NEW.type || ' tracking'
    );
END;

CREATE TRIGGER IF NOT EXISTS tracking_entries_update AFTER UPDATE ON tracking_entries
BEGIN
    UPDATE entities
    SET
        title = NEW.title,
        status = NEW.status,
        updated_at = NEW.updated_at,
        metadata = json_object(
            'source_file', NEW.source_file,
            'date', NEW.date,
            'type', NEW.type
        )
    WHERE id = NEW.id;

    UPDATE search_index
    SET
        title = NEW.title,
        content = NEW.content,
        tags = NEW.type || ' tracking'
    WHERE entity_id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS tracking_entries_delete AFTER DELETE ON tracking_entries
BEGIN
    DELETE FROM entities WHERE id = OLD.id;
    DELETE FROM search_index WHERE entity_id = OLD.id;
END;

-- Log the migration completion
INSERT INTO activity_log (id, action, domain, description, metadata) VALUES (
    'activity_' || hex(randomblob(8)),
    'migrate',
    'tracking',
    'Tracking domain migration completed',
    json_object(
        'migration', '005_tracking_domain',
        'tables_created', json_array('tracking_entries', 'tracking_data_points', 'tracking_types'),
        'default_types_created', 10,
        'indexes_created', 11,
        'triggers_created', 3
    )
);

-- +goose Down
DROP TRIGGER IF EXISTS tracking_entries_delete;
DROP TRIGGER IF EXISTS tracking_data_delete;
DROP TRIGGER IF EXISTS tracking_entries_update;
DROP TABLE IF EXISTS tracking_data_points;
DROP TABLE IF EXISTS tracking_entries;
DROP TABLE IF EXISTS tracking_types;
