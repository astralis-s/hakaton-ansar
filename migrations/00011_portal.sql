-- +goose Up
-- +goose StatementBegin
-- Client portal login credentials (one account per client; email is the login).
CREATE TABLE client_portal_accounts (
    client_id     uuid PRIMARY KEY REFERENCES clients(id) ON DELETE CASCADE,
    org_id        uuid        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email         text        NOT NULL UNIQUE,
    password_hash text        NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
-- One chat thread per client (client ↔ staff).
CREATE TABLE conversations (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          uuid        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    client_id       uuid        NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    created_at      timestamptz NOT NULL DEFAULT now(),
    last_message_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (org_id, client_id)
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_conversations_org ON conversations (org_id, last_message_at DESC);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE messages (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id uuid        NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    org_id          uuid        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    sender_kind     text        NOT NULL CHECK (sender_kind IN ('client', 'staff')),
    sender_id       uuid        NOT NULL,
    body            text        NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_messages_conversation ON messages (conversation_id, created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS messages;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS conversations;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS client_portal_accounts;
-- +goose StatementEnd
