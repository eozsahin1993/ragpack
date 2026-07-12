-- +goose Up
CREATE TABLE api_key_admin_grants (
    id            TEXT     PRIMARY KEY,
    api_key_id    TEXT     NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    resource_type TEXT     NOT NULL,
    permission    TEXT     NOT NULL CHECK (permission IN ('read', 'write', 'both')),
    created_at    DATETIME NOT NULL
);

CREATE INDEX idx_api_key_admin_grants_key ON api_key_admin_grants(api_key_id);

-- +goose Down
DROP TABLE api_key_admin_grants;
