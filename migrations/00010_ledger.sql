-- +goose Up
-- +goose StatementBegin
-- Manual business expenses (расходы): rent, repairs, logistics, … The cost of
-- goods sold is derived from contracts, not stored here.
CREATE TABLE expenses (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     uuid          NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    category   text          NOT NULL,
    amount     numeric(18,2) NOT NULL CHECK (amount > 0),
    note       text          NOT NULL DEFAULT '',
    spent_at   date          NOT NULL,
    created_at timestamptz   NOT NULL DEFAULT now()
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_expenses_org ON expenses (org_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS expenses;
-- +goose StatementEnd
