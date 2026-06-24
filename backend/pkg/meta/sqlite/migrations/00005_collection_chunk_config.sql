-- +goose Up
ALTER TABLE collections ADD COLUMN chunk_strategy TEXT;     -- NULL = use server default
ALTER TABLE collections ADD COLUMN chunk_size     INTEGER;  -- NULL = use server default
ALTER TABLE collections ADD COLUMN chunk_overlap  INTEGER;  -- NULL = use server default

-- +goose Down
ALTER TABLE collections DROP COLUMN chunk_strategy;
ALTER TABLE collections DROP COLUMN chunk_size;
ALTER TABLE collections DROP COLUMN chunk_overlap;
