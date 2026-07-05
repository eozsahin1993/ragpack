-- +goose Up
ALTER TABLE documents ADD COLUMN name TEXT;

-- +goose Down
ALTER TABLE documents DROP COLUMN name;
