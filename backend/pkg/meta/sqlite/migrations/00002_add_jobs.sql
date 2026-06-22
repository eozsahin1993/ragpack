-- +goose Up
CREATE TABLE jobs (
    id            TEXT PRIMARY KEY,
    collection_id TEXT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    file_uri      TEXT NOT NULL,
    mime_type     TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'pending',
    error         TEXT,
    executed_at   DATETIME,
    created_at    DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL
);

CREATE INDEX idx_jobs_collection ON jobs(collection_id);
CREATE INDEX idx_jobs_status ON jobs(status);

-- +goose Down
DROP INDEX IF EXISTS idx_jobs_status;
DROP INDEX IF EXISTS idx_jobs_collection;
DROP TABLE IF EXISTS jobs;
