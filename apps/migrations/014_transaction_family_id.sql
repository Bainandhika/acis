-- Allow income transactions to belong directly to a family primary balance.
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS family_id UUID REFERENCES families(id) ON DELETE CASCADE;

-- Keep existing wallet transactions associated with their family.

UPDATE transactions t
SET family_id = w.family_id
FROM wallets w
WHERE t.family_id IS NULL AND t.wallet_id = w.id;

CREATE INDEX IF NOT EXISTS idx_transactions_family_id ON transactions(family_id);