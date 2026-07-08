-- +goose Up
ALTER TABLE jobs ADD COLUMN metadata TEXT;

-- +goose Down
SELECT 1; -- SQLite does not support DROP COLUMN portably
