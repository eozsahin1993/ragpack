-- +goose Up
CREATE TABLE IF NOT EXISTS api_keys (
    id          TEXT    PRIMARY KEY,
    name        TEXT    NOT NULL,
    key_hash    TEXT    NOT NULL UNIQUE,
    key_hint    TEXT    NOT NULL,
    created_at  INTEGER NOT NULL,
    last_used_at INTEGER
);

-- +goose Down
DROP TABLE IF EXISTS api_keys;
