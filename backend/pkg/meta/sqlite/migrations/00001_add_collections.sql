-- +goose Up
CREATE TABLE collections (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    table_name  TEXT NOT NULL UNIQUE,
    embed_model TEXT NOT NULL,
    vector_dim  INTEGER NOT NULL,
    created_at  DATETIME NOT NULL,
    UNIQUE(name, embed_model)
);

-- +goose Down
DROP TABLE IF EXISTS collections;
