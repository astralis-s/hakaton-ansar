-- +goose Up
-- +goose StatementBegin
-- Client applications for a contract (заявки): the client picks a product and
-- states wishes; staff set terms and approve → a real contract is created.
CREATE TABLE contract_requests (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id               uuid          NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    client_id            uuid          NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    product_id           uuid          NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    desired_installments integer       NOT NULL,
    desired_down_payment numeric(18,2) NOT NULL DEFAULT 0,
    note                 text          NOT NULL DEFAULT '',
    status               text          NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    contract_id          uuid          REFERENCES contracts(id) ON DELETE SET NULL,
    created_at           timestamptz   NOT NULL DEFAULT now(),
    decided_at           timestamptz
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_contract_requests_org ON contract_requests (org_id, status, created_at DESC);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_contract_requests_client ON contract_requests (org_id, client_id, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS contract_requests;
-- +goose StatementEnd
