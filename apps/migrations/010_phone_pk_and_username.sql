-- Migration: 010_phone_pk_and_username.sql
-- Description: Make phone_number the sole primary key for users, add username column, drop email dependency
-- Date: 2026-08-16

-- +goose Up

-- 1. Add username column to users table if not exists
ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(100);

-- Backfill username from name or fallback
UPDATE users SET username = COALESCE(NULLIF(TRIM(name), ''), 'user_' || SUBSTRING(REPLACE(COALESCE(phone_number, '08123456789'), '+', '') FROM 4)) WHERE username IS NULL OR username = '';
ALTER TABLE users ALTER COLUMN username SET NOT NULL;

-- 2. Ensure users(id) has a unique constraint for FK references
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'users_id_unique'
    ) THEN
        ALTER TABLE users ADD CONSTRAINT users_id_unique UNIQUE (id);
    END IF;
END $$;

-- 3. Drop existing primary key (which was composite on (email, phone_number) or id)
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_pkey CASCADE;

-- 4. Re-establish foreign keys referencing users(id) via users_id_unique if dropped by cascade
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

-- 5. Set phone_number as the SOLE PRIMARY KEY on users
ALTER TABLE users ADD CONSTRAINT users_pkey PRIMARY KEY (phone_number);

-- 6. Allow email to be nullable if column exists, drop unique email constraint
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_unique;

-- 7. Add index for username lookup
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- +goose Down
-- Rollback script
-- ALTER TABLE users DROP CONSTRAINT IF EXISTS users_pkey;
-- ALTER TABLE users DROP COLUMN IF EXISTS username;
