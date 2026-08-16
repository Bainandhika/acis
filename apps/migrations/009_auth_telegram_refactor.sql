-- Migration: 009_auth_telegram_refactor.sql
-- Description: Refactor auth to Telegram Bot, add phone_number, composite PK (email, phone_number), uniqueness constraints, drop otp_codes
-- Date: 2026-08-16

-- +goose Up

-- 1. Add phone_number and telegram_chat_id to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone_number VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS telegram_chat_id BIGINT;

-- Backfill default placeholder phone_number for any existing rows without phone number
UPDATE users SET phone_number = '+6280000000000' WHERE phone_number IS NULL OR phone_number = '';
ALTER TABLE users ALTER COLUMN phone_number SET NOT NULL;

-- 2. Add unique constraint on users.id BEFORE dropping primary key, so foreign keys can reference it
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'users_id_unique'
    ) THEN
        ALTER TABLE users ADD CONSTRAINT users_id_unique UNIQUE (id);
    END IF;
END $$;

-- 3. Drop existing primary key on users using CASCADE to drop dependent foreign keys safely
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_pkey CASCADE;

-- 4. Re-establish foreign keys referencing users(id) via users_id_unique
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'families_created_by_fkey'
    ) THEN
        ALTER TABLE families ADD CONSTRAINT families_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'family_members_user_id_fkey'
    ) THEN
        ALTER TABLE family_members ADD CONSTRAINT family_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'wallets_created_by_fkey'
    ) THEN
        ALTER TABLE wallets ADD CONSTRAINT wallets_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'transactions_created_by_fkey'
    ) THEN
        ALTER TABLE transactions ADD CONSTRAINT transactions_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'proposals_proposed_by_fkey'
    ) THEN
        ALTER TABLE proposals ADD CONSTRAINT proposals_proposed_by_fkey FOREIGN KEY (proposed_by) REFERENCES users(id) ON DELETE SET NULL;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'proposals_reviewed_by_fkey'
    ) THEN
        ALTER TABLE proposals ADD CONSTRAINT proposals_reviewed_by_fkey FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- 5. Set composite primary key on (email, phone_number)
ALTER TABLE users ADD CONSTRAINT users_pkey PRIMARY KEY (email, phone_number);

-- 6. Enforce strict individual uniqueness so neither email nor phone number can be reused across different registrations
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'users_email_unique'
    ) THEN
        ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'users_phone_number_unique'
    ) THEN
        ALTER TABLE users ADD CONSTRAINT users_phone_number_unique UNIQUE (phone_number);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'users_telegram_chat_id_key'
    ) THEN
        ALTER TABLE users ADD CONSTRAINT users_telegram_chat_id_key UNIQUE (telegram_chat_id);
    END IF;
END $$;

-- 7. Drop obsolete email OTP table
DROP TABLE IF EXISTS otp_codes;

-- 8. Create performance indexes
CREATE INDEX IF NOT EXISTS idx_users_phone_number ON users(phone_number);
CREATE INDEX IF NOT EXISTS idx_users_telegram_chat_id ON users(telegram_chat_id);

-- +goose Down
-- Rollback safe script
-- ALTER TABLE users DROP CONSTRAINT IF EXISTS users_pkey;
-- ALTER TABLE users DROP CONSTRAINT IF EXISTS users_phone_number_unique;
-- ALTER TABLE users DROP CONSTRAINT IF EXISTS users_telegram_chat_id_key;
-- ALTER TABLE users ADD CONSTRAINT users_pkey PRIMARY KEY (id);
-- ALTER TABLE users DROP COLUMN IF EXISTS phone_number;
-- ALTER TABLE users DROP COLUMN IF EXISTS telegram_chat_id;
-- CREATE TABLE IF NOT EXISTS otp_codes (
--     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
--     email VARCHAR(255) NOT NULL,
--     code_hash VARCHAR(255) NOT NULL,
--     expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
--     is_used BOOLEAN DEFAULT FALSE,
--     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
-- );
