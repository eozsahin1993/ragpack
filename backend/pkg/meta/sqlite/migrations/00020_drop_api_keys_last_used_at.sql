-- +goose Up
-- last_used_at was written on every single authenticated request (see
-- ValidateAPIKey), making an otherwise pure-read auth path the one write on
-- SQLite's global writer lock that every request paid for. Not worth it for
-- a display-only "last used" timestamp — dropped rather than debounced.
ALTER TABLE api_keys DROP COLUMN last_used_at;

-- +goose Down
ALTER TABLE api_keys ADD COLUMN last_used_at DATETIME;
