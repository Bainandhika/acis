-- Migration: 015_supabase_auth_refactor.sql
-- Description: Refactor users into a Supabase Auth profile table (id = auth.users.id).
-- WARNING: Destructive. Acceptable only because the project currently holds seed/test data.
-- +goose Up

-- 1. Remove seed/test users so the FK to auth.users can be created cleanly.
DELETE FROM users;

-- 2. Drop legacy identity constraints (idempotent guards).
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_pkey CASCADE;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_id_unique CASCADE;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_unique CASCADE;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_phone_number_unique CASCADE;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_telegram_chat_id_key CASCADE;

-- 3. Drop legacy identity columns.
ALTER TABLE users DROP COLUMN IF EXISTS phone_number;
ALTER TABLE users DROP COLUMN IF EXISTS email;
ALTER TABLE users DROP COLUMN IF EXISTS google_id;

-- 4. id becomes the PK and mirrors auth.users.id.
ALTER TABLE users ALTER COLUMN id SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE users ADD CONSTRAINT users_id_fkey
    FOREIGN KEY (id) REFERENCES auth.users(id) ON DELETE CASCADE;

-- 4b. Re-establish foreign keys referencing users(id) if dropped by CASCADE
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'families_created_by_fkey') THEN
        ALTER TABLE families ADD CONSTRAINT families_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'family_members_user_id_fkey') THEN
        ALTER TABLE family_members ADD CONSTRAINT family_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'wallets_created_by_fkey') THEN
        ALTER TABLE wallets ADD CONSTRAINT wallets_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'transactions_created_by_fkey') THEN
        ALTER TABLE transactions ADD CONSTRAINT transactions_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'proposals_proposed_by_fkey') THEN
        ALTER TABLE proposals ADD CONSTRAINT proposals_proposed_by_fkey FOREIGN KEY (proposed_by) REFERENCES users(id) ON DELETE SET NULL;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'proposals_reviewed_by_fkey') THEN
        ALTER TABLE proposals ADD CONSTRAINT proposals_reviewed_by_fkey FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- 5. telegram_chat_id stays for bot linking: nullable, unique when set.
DROP INDEX IF EXISTS idx_users_telegram_chat_id;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_telegram_chat_id
    ON users (telegram_chat_id) WHERE telegram_chat_id IS NOT NULL;

-- 6. Backfill denormalized family_id on transactions (required by RLS policies).
UPDATE transactions t
SET family_id = w.family_id
FROM wallets w
WHERE t.wallet_id = w.id AND t.family_id IS NULL;

-- +goose Down
-- Rollback not supported.
