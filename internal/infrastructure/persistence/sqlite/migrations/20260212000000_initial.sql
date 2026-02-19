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

-- Helpful indexes
CREATE INDEX IF NOT EXISTS idx_finance_transactions_date ON finance_transactions(date);
CREATE INDEX IF NOT EXISTS idx_finance_transactions_account ON finance_transactions(account);
CREATE INDEX IF NOT EXISTS idx_finance_transactions_category ON finance_transactions(category);
CREATE INDEX IF NOT EXISTS idx_finance_transactions_excluded ON finance_transactions(excluded);
CREATE INDEX IF NOT EXISTS idx_finance_transactions_amount ON finance_transactions(amount);
CREATE INDEX IF NOT EXISTS idx_financial_accounts_name ON financial_accounts(name);

-- +goose Down
-- No down migration for initial schema
