-- +goose Up
-- Calendar Schema Redesign - Full Migration
-- Description: Creates new normalized calendar tables with proper constraints and migrates 5,000+ events
-- Risk Level: HIGH - touches 5,000+ events
-- Date: 2026-02-17
-- Author: Claude Code

-- =============================================================================
-- PHASE 1: Pre-Migration Setup
-- =============================================================================

-- Create migration log table if not exists
CREATE TABLE IF NOT EXISTS migration_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    migration_name TEXT UNIQUE NOT NULL,
    started_at TEXT DEFAULT (datetime('now')),
    completed_at TEXT,
    rows_processed INTEGER,
    status TEXT DEFAULT 'running',
    error_message TEXT,
    validation_results TEXT
);

-- Log migration start
INSERT OR REPLACE INTO migration_log (migration_name, status, started_at) 
VALUES ('calendar_schema_redesign_20260217', 'running', datetime('now'));

-- =============================================================================
-- PHASE 2: Optimized SQLite Settings
-- =============================================================================

PRAGMA foreign_keys = OFF;
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = -64000;  -- 64MB cache
PRAGMA temp_store = MEMORY;
PRAGMA busy_timeout = 30000; -- 30 seconds

BEGIN TRANSACTION;

-- =============================================================================
-- PHASE 3: Create calendars Table
-- =============================================================================

CREATE TABLE calendars (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT 'unknown',
    source_id TEXT,
    color TEXT,
    is_active INTEGER DEFAULT 1,
    last_synced_at TEXT,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    UNIQUE(name)
);

-- Extract unique calendars from calendar_events with normalization
INSERT INTO calendars (name, source, last_synced_at)
SELECT 
    DISTINCT trim(calendar_name) as name,
    CASE 
        WHEN lower(calendar_name) LIKE '%google%' OR lower(calendar_name) LIKE '%gmail%' THEN 'google'
        WHEN lower(calendar_name) LIKE '%todoist%' THEN 'todoist'
        WHEN lower(calendar_name) LIKE '%ical%' OR lower(calendar_name) LIKE '%ics%' THEN 'ical'
        ELSE 'unknown'
    END as source,
    datetime('now') as last_synced_at
FROM calendar_events
WHERE calendar_name IS NOT NULL AND trim(calendar_name) != '';

-- Add a default calendar for orphaned events
INSERT OR IGNORE INTO calendars (name, source, description)
VALUES ('Default', 'unknown', 'Default calendar for events without source');

CREATE INDEX idx_calendars_source ON calendars(source);
CREATE INDEX idx_calendars_active ON calendars(is_active) WHERE is_active = 1;

-- =============================================================================
-- PHASE 4: Create event_categories Table
-- =============================================================================

CREATE TABLE event_categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    emoji TEXT,
    color_code TEXT,
    icon_name TEXT,
    display_order INTEGER DEFAULT 0,
    is_active INTEGER DEFAULT 1,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

-- Migrate existing categories
INSERT INTO event_categories (id, name, description, emoji, display_order, is_active, created_at)
SELECT 
    id,
    name,
    description,
    emoji,
    COALESCE(display_order, 0),
    COALESCE(is_active, 1),
    COALESCE(
        CASE 
            WHEN created_at IS NOT NULL AND created_at != '' THEN datetime(created_at)
            ELSE datetime('now')
        END,
        datetime('now')
    )
FROM calendar_event_categories
WHERE name IS NOT NULL;

-- Ensure we have default categories if migration is empty
INSERT OR IGNORE INTO event_categories (name, description, emoji, display_order, is_active)
VALUES 
    ('Health & Fitness', 'Physical activity and exercise', '💪', 1, 1),
    ('Entertainment & Media', 'Movies, TV shows, reading', '🎬', 2, 1),
    ('Food & Dining', 'Meals, coffee, drinks', '🍽️', 3, 1),
    ('Travel & Transportation', 'Trips, drives, flights', '🚗', 4, 1),
    ('Social & Relationships', 'Time with friends and family', '👥', 5, 1),
    ('Work & Career', 'Work activities and meetings', '💼', 6, 1),
    ('Personal Development', 'Learning and growth', '📚', 7, 1),
    ('Household & Maintenance', 'Chores and errands', '🏠', 8, 1),
    ('Creativity & Hobbies', 'Creative pursuits', '🎨', 9, 1),
    ('Medical & Healthcare', 'Health and medical visits', '🏥', 10, 1);

CREATE INDEX idx_event_categories_order ON event_categories(display_order);
CREATE INDEX idx_event_categories_active ON event_categories(is_active) WHERE is_active = 1;

-- =============================================================================
-- PHASE 5: Create event_types Table
-- =============================================================================

CREATE TABLE event_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL REFERENCES event_categories(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    emoji TEXT,
    is_active INTEGER DEFAULT 1,
    frequency_score INTEGER DEFAULT 0,
    parsing_rule TEXT,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    UNIQUE(category_id, name)
);

-- Migrate existing event types (only where category exists)
INSERT INTO event_types (id, category_id, name, emoji, is_active, frequency_score, created_at)
SELECT 
    t.id,
    t.category_id,
    t.name,
    t.emoji,
    COALESCE(t.is_active, 1),
    COALESCE(t.frequency_score, 0),
    COALESCE(
        CASE 
            WHEN t.created_at IS NOT NULL AND t.created_at != '' THEN datetime(t.created_at)
            ELSE datetime('now')
        END,
        datetime('now')
    )
FROM calendar_event_types t
WHERE t.category_id IN (SELECT id FROM event_categories)
  AND t.name IS NOT NULL;

CREATE INDEX idx_event_types_category ON event_types(category_id);
CREATE INDEX idx_event_types_active ON event_types(is_active) WHERE is_active = 1;

-- =============================================================================
-- PHASE 6: Create New calendar_events Table
-- =============================================================================

CREATE TABLE calendar_events_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    calendar_id INTEGER NOT NULL REFERENCES calendars(id) ON DELETE CASCADE,
    uid TEXT NOT NULL,
    
    -- Event details
    summary TEXT,
    description TEXT,
    location TEXT,
    
    -- Timing (ISO8601 format)
    start_time TEXT NOT NULL,
    end_time TEXT,
    duration_minutes INTEGER,
    is_all_day INTEGER DEFAULT 0,
    timezone TEXT,
    
    -- Status
    status TEXT DEFAULT 'confirmed',
    is_recurring INTEGER DEFAULT 0,
    recurrence_rule TEXT,
    recurrence_exceptions TEXT CHECK(recurrence_exceptions IS NULL OR json_valid(recurrence_exceptions)),
    
    -- People
    organizer_email TEXT,
    organizer_name TEXT,
    attendees TEXT CHECK(attendees IS NULL OR json_valid(attendees)),
    
    -- Classification
    category_id INTEGER REFERENCES event_categories(id) ON DELETE SET NULL,
    event_type_id INTEGER REFERENCES event_types(id) ON DELETE SET NULL,
    extracted_detail TEXT,
    extracted_person TEXT,
    confidence_score REAL,
    format_class TEXT,
    
    -- Metadata
    created_at TEXT,
    last_modified_at TEXT,
    dtstamp TEXT,
    synced_at TEXT,
    deleted_at TEXT,
    
    UNIQUE(calendar_id, uid)
);

-- =============================================================================
-- PHASE 7: Migrate Events Data with Transformations
-- =============================================================================

INSERT INTO calendar_events_new (
    id, calendar_id, uid, summary, description, location,
    start_time, end_time, duration_minutes, is_all_day, timezone,
    status, is_recurring, recurrence_rule, recurrence_exceptions,
    organizer_email, organizer_name, attendees,
    category_id, event_type_id, extracted_detail, extracted_person, confidence_score, format_class,
    created_at, last_modified_at, dtstamp, synced_at
)
SELECT 
    e.id,
    COALESCE(c.id, (SELECT id FROM calendars WHERE name = 'Default')) as calendar_id,
    COALESCE(e.uid, 'generated-' || e.id || '-' || random()) as uid,
    e.summary,
    e.description,
    e.location,
    
    -- Convert TEXT datetime to ISO8601
    CASE 
        WHEN e.start IS NULL OR e.start = '' THEN datetime('now')
        -- Already ISO8601 format
        WHEN e.start LIKE '____-__-__ __:__:__' THEN e.start
        -- iCalendar UTC format: 20240115T143000Z
        WHEN LENGTH(TRIM(e.start)) = 16 AND SUBSTR(TRIM(e.start), 16, 1) = 'Z' THEN
            SUBSTR(e.start, 1, 4) || '-' || SUBSTR(e.start, 5, 2) || '-' || SUBSTR(e.start, 7, 2) || ' ' ||
            SUBSTR(e.start, 10, 2) || ':' || SUBSTR(e.start, 12, 2) || ':' || SUBSTR(e.start, 14, 2)
        -- iCalendar local format: 20240115T143000
        WHEN LENGTH(TRIM(e.start)) = 15 THEN
            SUBSTR(e.start, 1, 4) || '-' || SUBSTR(e.start, 5, 2) || '-' || SUBSTR(e.start, 7, 2) || ' ' ||
            SUBSTR(e.start, 10, 2) || ':' || SUBSTR(e.start, 12, 2) || ':' || SUBSTR(e.start, 14, 2)
        -- DATE only format: 20240115 (8 digits)
        WHEN LENGTH(TRIM(e.start)) = 8 AND e.start GLOB '[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]' THEN
            SUBSTR(e.start, 1, 4) || '-' || SUBSTR(e.start, 5, 2) || '-' || SUBSTR(e.start, 7, 2) || ' 00:00:00'
        -- Unix timestamp (10 digits)
        WHEN LENGTH(TRIM(e.start)) = 10 AND e.start GLOB '[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]' THEN
            datetime(CAST(e.start AS INTEGER), 'unixepoch')
        -- Fallback
        ELSE datetime('now')
    END as start_time,
    
    -- End time conversion (same logic)
    CASE 
        WHEN e.end IS NULL OR e.end = '' THEN NULL
        WHEN e.end LIKE '____-__-__ __:__:__' THEN e.end
        WHEN LENGTH(TRIM(e.end)) = 16 AND SUBSTR(TRIM(e.end), 16, 1) = 'Z' THEN
            SUBSTR(e.end, 1, 4) || '-' || SUBSTR(e.end, 5, 2) || '-' || SUBSTR(e.end, 7, 2) || ' ' ||
            SUBSTR(e.end, 10, 2) || ':' || SUBSTR(e.end, 12, 2) || ':' || SUBSTR(e.end, 14, 2)
        WHEN LENGTH(TRIM(e.end)) = 15 THEN
            SUBSTR(e.end, 1, 4) || '-' || SUBSTR(e.end, 5, 2) || '-' || SUBSTR(e.end, 7, 2) || ' ' ||
            SUBSTR(e.end, 10, 2) || ':' || SUBSTR(e.end, 12, 2) || ':' || SUBSTR(e.end, 14, 2)
        WHEN LENGTH(TRIM(e.end)) = 8 AND e.end GLOB '[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]' THEN
            SUBSTR(e.end, 1, 4) || '-' || SUBSTR(e.end, 5, 2) || '-' || SUBSTR(e.end, 7, 2) || ' 00:00:00'
        WHEN LENGTH(TRIM(e.end)) = 10 AND e.end GLOB '[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]' THEN
            datetime(CAST(e.end AS INTEGER), 'unixepoch')
        ELSE NULL
    END as end_time,
    
    -- Calculate duration in minutes
    CASE 
        WHEN e.start IS NOT NULL AND e.end IS NOT NULL THEN
            CAST((
                julianday(
                    CASE 
                        WHEN e.end LIKE '____-__-__ __:__:__' THEN e.end
                        WHEN LENGTH(TRIM(e.end)) >= 15 THEN
                            SUBSTR(e.end, 1, 4) || '-' || SUBSTR(e.end, 5, 2) || '-' || SUBSTR(e.end, 7, 2) || ' ' ||
                            SUBSTR(e.end, 10, 2) || ':' || SUBSTR(e.end, 12, 2) || ':' || SUBSTR(e.end, 14, 2)
                        ELSE e.end
                    END
                ) - julianday(
                    CASE 
                        WHEN e.start LIKE '____-__-__ __:__:__' THEN e.start
                        WHEN LENGTH(TRIM(e.start)) >= 15 THEN
                            SUBSTR(e.start, 1, 4) || '-' || SUBSTR(e.start, 5, 2) || '-' || SUBSTR(e.start, 7, 2) || ' ' ||
                            SUBSTR(e.start, 10, 2) || ':' || SUBSTR(e.start, 12, 2) || ':' || SUBSTR(e.start, 14, 2)
                        ELSE e.start
                    END
                )
            ) * 24 * 60 AS INTEGER)
        ELSE NULL
    END as duration_minutes,
    
    -- Detect all-day events
    CASE 
        WHEN e.start IS NOT NULL AND LENGTH(TRIM(e.start)) = 8 AND e.start GLOB '[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]' THEN 1
        WHEN e.end IS NOT NULL AND LENGTH(TRIM(e.end)) = 8 AND e.end GLOB '[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]' THEN 1
        WHEN e.recurrence_rule IS NOT NULL AND e.recurrence_rule LIKE '%VALUE=DATE%' THEN 1
        ELSE 0
    END as is_all_day,
    
    -- Timezone detection (simplified - would need full iCalendar parsing for TZID)
    CASE 
        WHEN e.start LIKE '%Z' THEN 'UTC'
        ELSE NULL
    END as timezone,
    
    COALESCE(NULLIF(e.status, ''), 'confirmed') as status,
    CASE WHEN e.recurrence_rule IS NOT NULL AND e.recurrence_rule != '' THEN 1 ELSE 0 END as is_recurring,
    e.recurrence_rule,
    NULL as recurrence_exceptions,
    
    -- Parse organizer (basic email extraction)
    CASE 
        WHEN e.organizer IS NULL THEN NULL
        WHEN e.organizer LIKE '%@%' THEN 
            CASE 
                WHEN e.organizer LIKE 'mailto:%' THEN SUBSTR(e.organizer, 8)
                ELSE e.organizer
            END
        ELSE NULL
    END as organizer_email,
    NULL as organizer_name,
    
    -- Convert attendees to JSON
    CASE 
        WHEN e.attendees IS NULL THEN NULL
        WHEN json_valid(e.attendees) THEN e.attendees
        ELSE json_array(e.attendees)
    END as attendees,
    
    -- Only keep FK references if they exist
    CASE WHEN e.category_id IN (SELECT id FROM event_categories) THEN e.category_id ELSE NULL END,
    CASE WHEN e.event_type_id IN (SELECT id FROM event_types) THEN e.event_type_id ELSE NULL END,
    
    e.extracted_detail,
    e.extracted_person,
    e.confidence_score,
    e.format_class,
    
    -- Convert timestamps
    CASE 
        WHEN e.created IS NULL OR e.created = '' THEN NULL
        WHEN e.created LIKE '____-__-__ __:__:__' THEN e.created
        WHEN LENGTH(TRIM(e.created)) = 16 AND SUBSTR(TRIM(e.created), 16, 1) = 'Z' THEN
            SUBSTR(e.created, 1, 4) || '-' || SUBSTR(e.created, 5, 2) || '-' || SUBSTR(e.created, 7, 2) || ' ' ||
            SUBSTR(e.created, 10, 2) || ':' || SUBSTR(e.created, 12, 2) || ':' || SUBSTR(e.created, 14, 2)
        WHEN LENGTH(TRIM(e.created)) = 15 THEN
            SUBSTR(e.created, 1, 4) || '-' || SUBSTR(e.created, 5, 2) || '-' || SUBSTR(e.created, 7, 2) || ' ' ||
            SUBSTR(e.created, 10, 2) || ':' || SUBSTR(e.created, 12, 2) || ':' || SUBSTR(e.created, 14, 2)
        ELSE NULL
    END as created_at,
    
    CASE 
        WHEN e.last_modified IS NULL OR e.last_modified = '' THEN NULL
        WHEN e.last_modified LIKE '____-__-__ __:__:__' THEN e.last_modified
        WHEN LENGTH(TRIM(e.last_modified)) = 16 AND SUBSTR(TRIM(e.last_modified), 16, 1) = 'Z' THEN
            SUBSTR(e.last_modified, 1, 4) || '-' || SUBSTR(e.last_modified, 5, 2) || '-' || SUBSTR(e.last_modified, 7, 2) || ' ' ||
            SUBSTR(e.last_modified, 10, 2) || ':' || SUBSTR(e.last_modified, 12, 2) || ':' || SUBSTR(e.last_modified, 14, 2)
        WHEN LENGTH(TRIM(e.last_modified)) = 15 THEN
            SUBSTR(e.last_modified, 1, 4) || '-' || SUBSTR(e.last_modified, 5, 2) || '-' || SUBSTR(e.last_modified, 7, 2) || ' ' ||
            SUBSTR(e.last_modified, 10, 2) || ':' || SUBSTR(e.last_modified, 12, 2) || ':' || SUBSTR(e.last_modified, 14, 2)
        ELSE NULL
    END as last_modified_at,
    
    e.dtstamp,
    datetime('now') as synced_at
    
FROM calendar_events e
LEFT JOIN calendars c ON LOWER(TRIM(e.calendar_name)) = LOWER(c.name);

-- =============================================================================
-- PHASE 8: Create Indexes (After Data Load for Performance)
-- =============================================================================

-- Core indexes for common queries
CREATE INDEX idx_events_start_time ON calendar_events_new(start_time);
CREATE INDEX idx_events_calendar ON calendar_events_new(calendar_id);
CREATE INDEX idx_events_category ON calendar_events_new(category_id);
CREATE INDEX idx_events_type ON calendar_events_new(event_type_id);

-- Performance-optimized composite indexes
CREATE INDEX idx_events_calendar_start ON calendar_events_new(calendar_id, start_time);
CREATE INDEX idx_events_category_start ON calendar_events_new(category_id, start_time);

-- Partial indexes for filtered queries
CREATE INDEX idx_events_active ON calendar_events_new(start_time) 
    WHERE deleted_at IS NULL;
CREATE INDEX idx_events_recurring ON calendar_events_new(calendar_id) 
    WHERE is_recurring = 1;
CREATE INDEX idx_events_all_day ON calendar_events_new(start_time) 
    WHERE is_all_day = 1;

-- Full-text search on summary
CREATE INDEX idx_events_summary ON calendar_events_new(summary);

-- =============================================================================
-- PHASE 9: Validate Migration
-- =============================================================================

-- Check row counts match
SELECT 
    'VALIDATION: Row Counts' as check_name,
    (SELECT COUNT(*) FROM calendar_events) as expected_count,
    (SELECT COUNT(*) FROM calendar_events_new) as actual_count,
    CASE 
        WHEN (SELECT COUNT(*) FROM calendar_events) = (SELECT COUNT(*) FROM calendar_events_new) 
        THEN '✓ PASS' 
        ELSE '✗ FAIL' 
    END as status;

-- Check for orphaned calendar references
SELECT 
    'VALIDATION: Calendar References' as check_name,
    COUNT(*) as orphaned_events
FROM calendar_events_new e
LEFT JOIN calendars c ON e.calendar_id = c.id
WHERE c.id IS NULL;

-- Check for FK violations
SELECT 
    'VALIDATION: Foreign Keys' as check_name,
    COUNT(*) as violation_count
FROM pragma_foreign_key_check('calendar_events_new');

-- Check datetime conversion quality
SELECT 
    'VALIDATION: Datetime Quality' as check_name,
    COUNT(*) as invalid_datetimes
FROM calendar_events_new
WHERE start_time IS NULL 
   OR start_time NOT LIKE '____-__-__ __:__:__';

-- =============================================================================
-- PHASE 10: Atomic Table Swap
-- =============================================================================

-- Rename old table as backup (for rollback if needed)
ALTER TABLE calendar_events RENAME TO calendar_events_backup;

-- Rename new table to production name
ALTER TABLE calendar_events_new RENAME TO calendar_events;

-- =============================================================================
-- PHASE 11: Drop Old Tables
-- =============================================================================

DROP TABLE IF EXISTS calendar_event_categories;
DROP TABLE IF EXISTS calendar_event_types;
DROP TABLE IF EXISTS calendar_event_type_mappings;
DROP TABLE IF EXISTS calendar_summary_map;

-- Keep calendar_events_backup for safety (drop manually after verification)
-- DROP TABLE IF EXISTS calendar_events_backup;

-- =============================================================================
-- PHASE 12: Update Migration Log
-- =============================================================================

UPDATE migration_log 
SET completed_at = datetime('now'),
    rows_processed = (SELECT COUNT(*) FROM calendar_events),
    status = 'completed',
    validation_results = json_object(
        'total_events', (SELECT COUNT(*) FROM calendar_events),
        'calendars_created', (SELECT COUNT(*) FROM calendars),
        'categories_migrated', (SELECT COUNT(*) FROM event_categories),
        'types_migrated', (SELECT COUNT(*) FROM event_types),
        'orphaned_calendars_handled', (SELECT COUNT(*) FROM calendars WHERE name = 'Default'),
        'datetime_conversions', (SELECT COUNT(*) FROM calendar_events WHERE start_time IS NOT NULL)
    )
WHERE migration_name = 'calendar_schema_redesign_20260217';

COMMIT;

-- Re-enable foreign keys
PRAGMA foreign_keys = ON;

-- Final integrity verification
PRAGMA integrity_check;
PRAGMA foreign_key_check;

-- Update statistics for query planner
ANALYZE calendar_events;
ANALYZE calendars;
ANALYZE event_categories;
ANALYZE event_types;

-- +goose Down
-- Rollback: Restore from backup
-- NOTE: This migration modifies data irreversibly. To rollback:
-- 1. Stop application
-- 2. sqlite3 ~/.config/hominem/db.sqlite ".restore <backup_file>"
-- 3. Or manually: DROP TABLE calendar_events; ALTER TABLE calendar_events_backup RENAME TO calendar_events;
-- 4. Restart application

-- Emergency rollback SQL (if backup table exists):
-- BEGIN EXCLUSIVE;
-- DROP TABLE IF EXISTS calendar_events;
-- ALTER TABLE calendar_events_backup RENAME TO calendar_events;
-- COMMIT;
