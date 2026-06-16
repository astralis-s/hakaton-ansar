-- +goose Up
-- +goose StatementBegin
-- Telegram users of the support bot. chat_id is the Telegram private-chat id.
-- A user is linked to a CRM client once registration (full name + phone) is done;
-- until then client_id is NULL. The chat itself lives in the portal `messages`
-- table (the bot relays into the existing staff chat), so this table only holds
-- the identity/registration mapping.
CREATE TABLE telegram_users (
    chat_id    bigint PRIMARY KEY,
    org_id     uuid        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    client_id  uuid            NULL REFERENCES clients(id) ON DELETE SET NULL,
    username   text        NOT NULL DEFAULT '',
    full_name  text        NOT NULL DEFAULT '',
    phone      text        NOT NULL DEFAULT '',
    state      text        NOT NULL CHECK (state IN ('awaiting_name', 'awaiting_phone', 'registered')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
-- One Telegram chat per client (when linked).
CREATE UNIQUE INDEX idx_telegram_users_client ON telegram_users (org_id, client_id) WHERE client_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS telegram_users;
-- +goose StatementEnd
