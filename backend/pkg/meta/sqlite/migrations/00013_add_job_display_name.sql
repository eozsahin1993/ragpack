-- +goose Up
ALTER TABLE jobs ADD COLUMN display_name TEXT;

-- +goose Down
ALTER TABLE jobs DROP COLUMN display_name;
