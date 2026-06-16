-- +goose Up
-- +goose StatementBegin
ALTER TABLE reminders
    ADD COLUMN status text NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'completed', 'cancelled')),
    ADD COLUMN completed_at timestamptz,
    ADD COLUMN cancelled_at timestamptz;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_reminders_org_status ON reminders (org_id, status, scheduled_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_reminders_org_status;
ALTER TABLE reminders
    DROP COLUMN IF EXISTS cancelled_at,
    DROP COLUMN IF EXISTS completed_at,
    DROP COLUMN IF EXISTS status;
-- +goose StatementEnd
