-- +goose Up
-- +goose StatementBegin
CREATE TABLE reminders (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id           uuid        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    type             text        NOT NULL CHECK (type IN ('call', 'delivery', 'payment_followup')),
    client_id        uuid        REFERENCES clients(id),
    contract_id      uuid        REFERENCES contracts(id),
    note             text        NOT NULL DEFAULT '',
    desired_at       timestamptz NOT NULL,
    duration_minutes integer     NOT NULL DEFAULT 0,
    scheduled_at     timestamptz NOT NULL,
    was_shifted      boolean     NOT NULL DEFAULT false,
    reason           text        NOT NULL DEFAULT '',
    created_at       timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_reminders_org ON reminders (org_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS reminders;
-- +goose StatementEnd
