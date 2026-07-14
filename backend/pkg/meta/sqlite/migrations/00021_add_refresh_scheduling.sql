-- +goose Up
ALTER TABLE collections ADD COLUMN refresh_enabled INTEGER NOT NULL DEFAULT 0;
ALTER TABLE collections ADD COLUMN refresh_interval_seconds INTEGER;
ALTER TABLE collections ADD COLUMN last_auto_refresh_at DATETIME; -- NULL = never checked

ALTER TABLE documents ADD COLUMN last_etag TEXT; -- NULL = source sent no ETag, or never checked

-- +goose Down
ALTER TABLE documents DROP COLUMN last_etag;
ALTER TABLE collections DROP COLUMN last_auto_refresh_at;
ALTER TABLE collections DROP COLUMN refresh_interval_seconds;
ALTER TABLE collections DROP COLUMN refresh_enabled;
