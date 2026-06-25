-- +goose Up
ALTER TABLE jobs ADD COLUMN intent TEXT    NOT NULL DEFAULT 'ingest';
ALTER TABLE jobs ADD COLUMN force  INTEGER NOT NULL DEFAULT 0;

-- Enforce one document per source per collection.
-- If this fails there are duplicate file_uri rows — clean them up first.
CREATE UNIQUE INDEX idx_documents_collection_file_uri
    ON documents(collection_id, file_uri);

-- +goose Down
DROP INDEX IF EXISTS idx_documents_collection_file_uri;
ALTER TABLE jobs DROP COLUMN intent;
ALTER TABLE jobs DROP COLUMN force;
