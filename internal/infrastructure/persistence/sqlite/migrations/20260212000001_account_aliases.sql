-- +goose Up
-- Migration: 002_account_aliases
-- Description: Creates table for account alias management
-- Created: February 2026

-- Create account aliases table
CREATE TABLE IF NOT EXISTS account_aliases (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    alias TEXT NOT NULL UNIQUE,
    canonical_name TEXT NOT NULL,
    account_id INTEGER,
    confidence_score REAL DEFAULT 1.0,
    validation_count INTEGER DEFAULT 0,
    last_seen_at TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (account_id) REFERENCES financial_accounts(id)
);

-- Create indexes for efficient lookups
CREATE INDEX IF NOT EXISTS idx_aliases_alias ON account_aliases(alias);
CREATE INDEX IF NOT EXISTS idx_aliases_canonical ON account_aliases(canonical_name);

-- Insert validated aliases from existing data analysis
-- These are the confirmed mappings from the current codebase
INSERT OR IGNORE INTO account_aliases (alias, canonical_name, confidence_score) VALUES
    ('Platinum Card®', 'American Express Platinum', 1.0),
    ('Rewards Checking', 'American Express Rewards Checking', 1.0),
    ('High Yield Savings Acct', 'American Express Savings', 1.0),
    ('Captial One 360 Checking', 'Capital One 360 Checking', 1.0),
    ('Captial One 360 Performance Savings', 'Capital One 360 Performance Savings', 1.0),
    ('Captial One Venture One', 'Capital One Venture One', 1.0),
    ('Captial One Quicksilver', 'Capital One Quicksilver', 1.0),
    ('Captial One 360 Money Market', 'Capital One 360 Money Market', 1.0),
    ('Captial One Savor', 'Capital One Savor', 1.0),
    ('Streamyard Inc. 401(k) Plan', 'Bending Spoons Us 401k Plan', 1.0),
    -- Comma-separated aliases (first part only after normalization)
    ('American Express Rewards Checking,Rewards Checking', 'American Express Rewards Checking', 1.0),
    ('Capital One 360 Checking,Capital One 360 Checking', 'Capital One 360 Checking', 1.0),
    ('High Yield Savings Acct,American Express Savings', 'American Express Savings', 1.0),
    ('Captial One 360 Money Market,Captial One 360 Money Market', 'Capital One 360 Money Market', 1.0),
    ('Captial One 360 Performance Savings,Captial One 360 Performance Savings', 'Capital One 360 Performance Savings', 1.0),
    ('American Express Platinum,Platinum Card®', 'American Express Platinum', 1.0),
    ('Captial One Quicksilver,Captial One Quicksilver', 'Capital One Quicksilver', 1.0),
    ('Captial One Venture One,Captial One Venture One', 'Capital One Venture One', 1.0);

-- +goose Down
DROP TABLE IF EXISTS account_aliases;
