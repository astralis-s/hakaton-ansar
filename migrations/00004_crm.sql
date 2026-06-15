-- +goose Up
-- +goose StatementBegin
CREATE TABLE clients (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     uuid        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    full_name  text        NOT NULL,
    phone      text        NOT NULL DEFAULT '',
    document   text        NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_clients_org ON clients (org_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS clients;
-- +goose StatementEnd
