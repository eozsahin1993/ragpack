-- +goose Up
-- SQLite doesn't support ALTER COLUMN, so recreate documents with nullable job_id + ON DELETE SET NULL
CREATE TABLE documents_new (
    id            TEXT     NOT NULL PRIMARY KEY,
    collection_id TEXT     NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    job_id        TEXT     UNIQUE REFERENCES jobs(id) ON DELETE SET NULL,
    file_uri      TEXT     NOT NULL,
    mime_type     TEXT     NOT NULL,
    external_id   TEXT,
    extra_json    TEXT,
    chunk_count   INTEGER  NOT NULL DEFAULT 0,
    status        TEXT     NOT NULL DEFAULT 'ingesting',
    error         TEXT,
    created_at    DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL
);

INSERT INTO documents_new SELECT * FROM documents;
DROP TABLE documents;
ALTER TABLE documents_new RENAME TO documents;
CREATE INDEX idx_documents_collection_id ON documents(collection_id);

-- +goose Down
CREATE TABLE documents_old (
    id            TEXT     NOT NULL PRIMARY KEY,
    collection_id TEXT     NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    job_id        TEXT     NOT NULL UNIQUE REFERENCES jobs(id),
    file_uri      TEXT     NOT NULL,
    mime_type     TEXT     NOT NULL,
    external_id   TEXT,
    extra_json    TEXT,
    chunk_count   INTEGER  NOT NULL DEFAULT 0,
    status        TEXT     NOT NULL DEFAULT 'ingesting',
    error         TEXT,
    created_at    DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL
);

INSERT INTO documents_old SELECT * FROM documents WHERE job_id IS NOT NULL;
DROP TABLE documents;
ALTER TABLE documents_old RENAME TO documents;
CREATE INDEX idx_documents_collection_id ON documents(collection_id);
