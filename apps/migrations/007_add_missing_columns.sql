-- Migration: 007_add_missing_columns.sql
-- Description: Add category to transactions, title to proposals, telegram_chat_id to families
-- Date: 2026-08-12

-- Add category to transactions (for envelope categorization)
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS category VARCHAR(100);

-- Add title to proposals (short summary of the expense request)
ALTER TABLE proposals ADD COLUMN IF NOT EXISTS title VARCHAR(255);

-- Add telegram_chat_id to families (for bot routing)
ALTER TABLE families ADD COLUMN IF NOT EXISTS telegram_chat_id BIGINT;
