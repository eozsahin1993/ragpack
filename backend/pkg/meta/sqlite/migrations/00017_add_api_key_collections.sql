-- +goose Up
CREATE TABLE api_key_collections (
    id            TEXT     PRIMARY KEY,
    api_key_id    TEXT     NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    collection_id TEXT     REFERENCES collections(id) ON DELETE CASCADE,
    permission    TEXT     NOT NULL CHECK (permission IN ('read', 'write', 'both')),
    created_at    DATETIME NOT NULL
);

CREATE INDEX idx_api_key_collections_key ON api_key_collections(api_key_id);

-- +goose Down
DROP TABLE api_key_collections;
