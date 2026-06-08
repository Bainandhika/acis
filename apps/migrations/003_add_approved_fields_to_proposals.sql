-- +goose Up
ALTER TABLE proposals
ADD COLUMN approved_by UUID REFERENCES users(id),
ADD COLUMN approved_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE proposals
DROP COLUMN approved_by,
DROP COLUMN approved_at;