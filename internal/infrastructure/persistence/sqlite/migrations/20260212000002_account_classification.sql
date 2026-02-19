-- +goose Up
-- Migration: 003_account_classification
-- Description: Ensures all accounts are properly classified and adds any missing accounts from transactions
-- Created: February 2026

-- Add missing accounts from transactions that aren't in financial_accounts
INSERT OR IGNORE INTO financial_accounts (name, type, active)
SELECT DISTINCT 
    account,
    CASE 
        WHEN account LIKE '%Checking%' OR account LIKE '%Current%' THEN 'CHECKING'
        WHEN account LIKE '%Savings%' OR account LIKE '%Money Market%' THEN 'SAVINGS'
        WHEN account LIKE '%Card%' OR account LIKE '%Platinum%' OR account LIKE '%Sapphire%' OR account LIKE '%Gold%' THEN 'CREDIT'
        WHEN account LIKE '%401k%' OR account LIKE '%IRA%' OR account LIKE '%Investment%' THEN 'INVESTMENTS'
        WHEN account LIKE '%Cash%' THEN 'CASH'
        ELSE 'UNKNOWN'
    END,
    1
FROM finance_transactions
WHERE account IS NOT NULL 
  AND account != ''
  AND account NOT IN (SELECT name FROM financial_accounts);

-- Update existing accounts with proper types based on name heuristics
UPDATE financial_accounts SET type = 'CREDIT' WHERE name LIKE '%Platinum%' AND (type IS NULL OR type = 'UNKNOWN');
UPDATE financial_accounts SET type = 'CREDIT' WHERE name LIKE '%Gold%' AND (type IS NULL OR type = 'UNKNOWN');
UPDATE financial_accounts SET type = 'CREDIT' WHERE name LIKE '%Sapphire%' AND (type IS NULL OR type = 'UNKNOWN');
UPDATE financial_accounts SET type = 'CREDIT' WHERE name LIKE '%Venture%' AND (type IS NULL OR type = 'UNKNOWN');
UPDATE financial_accounts SET type = 'CREDIT' WHERE name LIKE '%Quicksilver%' AND (type IS NULL OR type = 'UNKNOWN');
UPDATE financial_accounts SET type = 'CREDIT' WHERE name LIKE '%Savor%' AND (type IS NULL OR type = 'UNKNOWN');
UPDATE financial_accounts SET type = 'CREDIT' WHERE name LIKE '%Rewards+%' AND (type IS NULL OR type = 'UNKNOWN');
UPDATE financial_accounts SET type = 'CREDIT' WHERE name LIKE '%Double Cash%' AND (type IS NULL OR type = 'UNKNOWN');
UPDATE financial_accounts SET type = 'CREDIT' WHERE name LIKE '%View%' AND (type IS NULL OR type = 'UNKNOWN');
UPDATE financial_accounts SET type = 'CREDIT' WHERE name LIKE '%Card%' AND (type IS NULL OR type = 'UNKNOWN');

UPDATE financial_accounts SET type = 'CHECKING' WHERE name LIKE '%Checking%' AND (type IS NULL OR type = 'UNKNOWN');
UPDATE financial_accounts SET type = 'CHECKING' WHERE name LIKE '%Current%' AND (type IS NULL OR type = 'UNKNOWN');

UPDATE financial_accounts SET type = 'SAVINGS' WHERE name LIKE '%Savings%' AND (type IS NULL OR type = 'UNKNOWN');
UPDATE financial_accounts SET type = 'SAVINGS' WHERE name LIKE '%Money Market%' AND (type IS NULL OR type = 'UNKNOWN');

UPDATE financial_accounts SET type = 'INVESTMENTS' WHERE name LIKE '%401k%' AND (type IS NULL OR type = 'UNKNOWN');
UPDATE financial_accounts SET type = 'INVESTMENTS' WHERE name LIKE '%IRA%' AND (type IS NULL OR type = 'UNKNOWN');

UPDATE financial_accounts SET type = 'CASH' WHERE name LIKE '%Cash%' AND (type IS NULL OR type = 'UNKNOWN');
UPDATE financial_accounts SET type = 'CASH' WHERE name LIKE '%Pay%' AND (type IS NULL OR type = 'UNKNOWN');

-- Ensure active flag is set
UPDATE financial_accounts SET active = 1 WHERE active IS NULL;

-- +goose Down
-- No down migration for data classification
