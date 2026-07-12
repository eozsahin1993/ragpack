-- +goose Up
-- SQLite has no ALTER COLUMN, so converting created_at/last_used_at from
-- unix-seconds INTEGER to DATETIME (matching every other table) means a
-- rebuild: create the new shape, copy data (converting the timestamps),
-- drop the old table, rename. api_key_collections/api_key_admin_grants
-- reference api_keys(id) by value (TEXT UUID), which is untouched by this
-- rebuild, so those FKs stay valid across the swap.
CREATE TABLE api_keys_new (
    id           TEXT     PRIMARY KEY,
    name         TEXT     NOT NULL,
    key_hash     TEXT     NOT NULL UNIQUE,
    key_hint     TEXT     NOT NULL,
    created_at   DATETIME NOT NULL,
    last_used_at DATETIME
);

INSERT INTO api_keys_new (id, name, key_hash, key_hint, created_at, last_used_at)
SELECT id, name, key_hash, key_hint,
       datetime(created_at, 'unixepoch'),
       CASE WHEN last_used_at IS NULL THEN NULL ELSE datetime(last_used_at, 'unixepoch') END
FROM api_keys;

DROP TABLE api_keys;
ALTER TABLE api_keys_new RENAME TO api_keys;

-- +goose Down
CREATE TABLE api_keys_old (
    id           TEXT    PRIMARY KEY,
    name         TEXT    NOT NULL,
    key_hash     TEXT    NOT NULL UNIQUE,
    key_hint     TEXT    NOT NULL,
    created_at   INTEGER NOT NULL,
    last_used_at INTEGER
);

INSERT INTO api_keys_old (id, name, key_hash, key_hint, created_at, last_used_at)
SELECT id, name, key_hash, key_hint,
       strftime('%s', created_at),
       CASE WHEN last_used_at IS NULL THEN NULL ELSE strftime('%s', last_used_at) END
FROM api_keys;

DROP TABLE api_keys;
ALTER TABLE api_keys_old RENAME TO api_keys;
