-- +goose Up
-- +goose StatementBegin
-- The sadaqa (charity) registry feature was removed from the product.
DROP TABLE IF EXISTS charity_entries;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE charity_entries (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      uuid           NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    contract_id uuid           NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    client_id   uuid           NOT NULL REFERENCES clients(id),
    amount      numeric(18, 2) NOT NULL,
    status      text           NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'transferred')),
    note        text           NOT NULL DEFAULT '',
    created_by  uuid           NOT NULL REFERENCES users(id),
    created_at  timestamptz    NOT NULL DEFAULT now()
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_charity_org ON charity_entries (org_id);
-- +goose StatementEnd
