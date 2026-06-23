-- +goose Up
CREATE TABLE documents (
    id            TEXT    NOT NULL PRIMARY KEY,
    collection_id TEXT    NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    job_id        TEXT    NOT NULL UNIQUE REFERENCES jobs(id),
    file_uri      TEXT    NOT NULL,
    mime_type     TEXT    NOT NULL,
    external_id   TEXT,
    extra_json    TEXT,
    chunk_count   INTEGER NOT NULL DEFAULT 0,
    status        TEXT    NOT NULL DEFAULT 'ingesting',
    error         TEXT,
    created_at    DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL
);

CREATE INDEX idx_documents_collection_id ON documents(collection_id);

-- +goose Down
DROP TABLE documents;
