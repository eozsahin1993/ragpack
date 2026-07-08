-- +goose Up
CREATE TABLE collection_metadata_fields (
    id            TEXT PRIMARY KEY,
    collection_id TEXT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    type          TEXT NOT NULL,
    slot          INTEGER NOT NULL,
    created_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    UNIQUE(collection_id, name),
    UNIQUE(collection_id, type, slot)
);

-- +goose Down
DROP TABLE collection_metadata_fields;
