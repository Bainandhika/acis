-- Migration: 015_supabase_auth_refactor.sql
-- Description: Refactor users into a Supabase Auth profile table (id = auth.users.id).
-- WARNING: Destructive. Acceptable only because the project currently holds seed/test data.
-- +goose Up

-- 1. Remove seed/test users so the FK to auth.users can be created cleanly.
DELETE FROM users;

-- 2. Drop legacy identity constraints (idempotent guards).
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_pkey;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_id_unique;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_unique;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_phone_number_unique;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_telegram_chat_id_key;

-- 3. Drop legacy identity columns.
ALTER TABLE users DROP COLUMN IF EXISTS phone_number;
ALTER TABLE users DROP COLUMN IF EXISTS email;
ALTER TABLE users DROP COLUMN IF EXISTS google_id;

-- 4. id becomes the PK and mirrors auth.users.id.
ALTER TABLE users ALTER COLUMN id SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE users ADD CONSTRAINT users_id_fkey
    FOREIGN KEY (id) REFERENCES auth.users(id) ON DELETE CASCADE;

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
