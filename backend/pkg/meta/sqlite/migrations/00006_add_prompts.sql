-- +goose Up
CREATE TABLE prompts (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    slug       TEXT NOT NULL UNIQUE,
    content    TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS prompts;
