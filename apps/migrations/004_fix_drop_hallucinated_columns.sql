-- +goose Up
-- Drop the hallucinated columns that don't match our domain model
ALTER TABLE proposals DROP COLUMN IF EXISTS approved_by;
ALTER TABLE proposals DROP COLUMN IF EXISTS approved_at;

-- +goose Down
-- Rollback: recreate the wrong columns just in case (though we shouldn't need them)
ALTER TABLE proposals ADD COLUMN IF NOT EXISTS approved_by UUID REFERENCES users(id);
ALTER TABLE proposals ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ;