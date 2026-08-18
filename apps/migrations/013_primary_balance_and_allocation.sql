-- Migration 013: Add primary_balance to families and support allocation
ALTER TABLE families 
ADD COLUMN IF NOT EXISTS primary_balance NUMERIC(15, 2) NOT NULL DEFAULT 0.00;

-- Optional: If transactions table has a check constraint on type, update it
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.constraint_column_usage 
        WHERE table_name = 'transactions' AND constraint_name LIKE '%type%'
    ) THEN
        ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_type_check;
        ALTER TABLE transactions ADD CONSTRAINT transactions_type_check CHECK (type IN ('income', 'expense', 'allocation'));
    END IF;
END $$;

-- Create index for monthly queries on transactions joined with wallets
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at);
