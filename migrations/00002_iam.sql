-- +goose Up
-- +goose StatementBegin
CREATE TABLE organizations (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name       text        NOT NULL,
    currency   text        NOT NULL DEFAULT 'RUB',
    created_at timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        uuid        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    full_name     text        NOT NULL,
    email         text        NOT NULL,
    password_hash text        NOT NULL,
    role          text        NOT NULL CHECK (role IN ('owner', 'manager')),
    created_at    timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT users_email_unique UNIQUE (email)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_users_org ON users (org_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE api_keys (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     uuid        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name       text        NOT NULL,
    prefix     text        NOT NULL,
    key_hash   text        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    revoked_at timestamptz,
    CONSTRAINT api_keys_hash_unique UNIQUE (key_hash)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_api_keys_org ON api_keys (org_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS api_keys;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS organizations;
-- +goose StatementEnd
