-- +goose Up
ALTER TABLE jobs ADD COLUMN extra_json TEXT;

-- +goose Down
SELECT 1; -- SQLite does not support DROP COLUMN in older versions
