-- +goose Up
-- Migration: 001_initial
-- Description: Creates core financial tables
-- Created: February 2026

-- Core transaction data
CREATE TABLE IF NOT EXISTS finance_transactions (
    id INTEGER PRIMARY KEY,
    date TEXT NOT NULL,
    name TEXT NOT NULL,
    amount REAL NOT NULL,
    status TEXT NOT NULL,
    category TEXT NOT NULL,
    parent_category TEXT NOT NULL,
    excluded INTEGER DEFAULT 0,
    tags TEXT,
    type TEXT NOT NULL,
    account TEXT NOT NULL,
    account_mask TEXT,
    note TEXT,
    recurring INTEGER,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- Account metadata
CREATE TABLE IF NOT EXISTS financial_accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    type TEXT,
    credit_limit REAL,
    active INTEGER
);

-- Net worth history
CREATE TABLE IF NOT EXISTS runway (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT,
    available_funds REAL,
    weight REAL
);

-- Recurring expenses
CREATE TABLE IF NOT EXISTS finance_expenses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    payee TEXT,
    monthly_cost REAL,
    type TEXT,
    billing_period TEXT,
    situation TEXT,
    year INTEGER,
    category TEXT,
    start_date TEXT,
    end_date TEXT,
    annual_cost REAL
);

-- Calendar events (added to fulfill migration dependencies)
CREATE TABLE IF NOT EXISTS calendar_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT,
    calendar_id INTEGER,
    calendar_name TEXT,
    summary TEXT,
    description TEXT,
    location TEXT,
    start TEXT,
    end TEXT,
    start_time TEXT,
    end_time TEXT,
    is_all_day INTEGER DEFAULT 0,
    is_recurring INTEGER DEFAULT 0,
    recurrence_rule TEXT,
    status TEXT DEFAULT 'confirmed',
    organizer TEXT,
    attendees TEXT,
    ical_uid TEXT,
    category_id INTEGER,
    event_type_id INTEGER,
    extracted_detail TEXT,
    extracted_person TEXT,
    confidence_score REAL,
    format_class TEXT,
    duration_minutes INTEGER,
    created TEXT,
    last_modified TEXT,
    dtstamp TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP,
    deleted_at TEXT
);

-- Calendar categories (Legacy support)
CREATE TABLE IF NOT EXISTS calendar_event_categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    emoji TEXT,
    display_order INTEGER,
    is_active INTEGER,
    created_at TEXT
);

-- Calendar types (Legacy support)
CREATE TABLE IF NOT EXISTS calendar_event_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER,
    name TEXT NOT NULL,
    description TEXT,
    emoji TEXT,
    is_active INTEGER,
    display_order INTEGER,
    frequency_score INTEGER,
    created_at TEXT
);

-- Helpful indexes
CREATE INDEX IF NOT EXISTS idx_finance_transactions_date ON finance_transactions(date);
CREATE INDEX IF NOT EXISTS idx_finance_transactions_account ON finance_transactions(account);
CREATE INDEX IF NOT EXISTS idx_finance_transactions_category ON finance_transactions(category);
CREATE INDEX IF NOT EXISTS idx_finance_transactions_excluded ON finance_transactions(excluded);
CREATE INDEX IF NOT EXISTS idx_finance_transactions_amount ON finance_transactions(amount);
CREATE INDEX IF NOT EXISTS idx_financial_accounts_name ON financial_accounts(name);

-- +goose Down
-- No down migration for initial schema
