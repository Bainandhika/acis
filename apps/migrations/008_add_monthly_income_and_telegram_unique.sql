-- Migration: 008_add_monthly_income_and_telegram_unique.sql
-- Description: Add monthly_income to families and unique constraint on telegram_chat_id
-- Date: 2026-08-12

ALTER TABLE families ADD COLUMN IF NOT EXISTS monthly_income NUMERIC(15, 2) DEFAULT 0.00;

DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'families_telegram_chat_id_key'
    ) THEN
        ALTER TABLE families ADD CONSTRAINT families_telegram_chat_id_key UNIQUE (telegram_chat_id);
    END IF;
END $$;
