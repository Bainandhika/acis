-- Migration: 012_wallet_short_id_and_no_email.sql
-- Description: Add short_id to wallets (invite_code-N), drop users.email, drop transactions.category, extend proposals

-- +goose Up

-- 1. Add short_id column to wallets (nullable first, then backfill)
ALTER TABLE wallets ADD COLUMN IF NOT EXISTS short_id VARCHAR(20);

-- 2. Backfill short_id for existing wallets using invite_code + row_number per family
WITH ranked AS (
  SELECT w.id,
         f.invite_code || '-' || ROW_NUMBER() OVER (PARTITION BY w.family_id ORDER BY w.created_at) AS sid
  FROM wallets w
  JOIN families f ON w.family_id = f.id
)
UPDATE wallets w SET short_id = ranked.sid
FROM ranked WHERE w.id = ranked.id;

ALTER TABLE wallets ALTER COLUMN short_id SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_wallets_short_id ON wallets(short_id);

-- 3. Add wallet_counter to families for auto-increment short_id generation
ALTER TABLE families ADD COLUMN IF NOT EXISTS wallet_counter INTEGER NOT NULL DEFAULT 0;

-- 4. Sync counter to current max per family
UPDATE families f
SET wallet_counter = (SELECT COUNT(*) FROM wallets w WHERE w.family_id = f.id);

-- 5. Drop category column from transactions
ALTER TABLE transactions DROP COLUMN IF EXISTS category;

-- 6. Drop request_type, extend proposals for pending requests
ALTER TABLE proposals ADD COLUMN IF NOT EXISTS request_type VARCHAR(50) NOT NULL DEFAULT 'add_transaction';
ALTER TABLE proposals ADD COLUMN IF NOT EXISTS target_transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL;
ALTER TABLE proposals ADD COLUMN IF NOT EXISTS payload JSONB;

-- 7. Drop users.email (already nullable after migration 010; drop the column fully)
ALTER TABLE users DROP COLUMN IF EXISTS email;

-- +goose Down
ALTER TABLE wallets DROP COLUMN IF EXISTS short_id;
ALTER TABLE families DROP COLUMN IF EXISTS wallet_counter;
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS category VARCHAR(100);
ALTER TABLE proposals DROP COLUMN IF EXISTS request_type;
ALTER TABLE proposals DROP COLUMN IF EXISTS target_transaction_id;
ALTER TABLE proposals DROP COLUMN IF EXISTS payload;
ALTER TABLE users ADD COLUMN IF NOT EXISTS email VARCHAR(255);
