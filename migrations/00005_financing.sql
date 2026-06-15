-- +goose Up
-- +goose StatementBegin
CREATE TABLE contracts (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id             uuid           NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    client_id          uuid           NOT NULL REFERENCES clients(id),
    product_id         uuid           NOT NULL REFERENCES products(id),
    cost_price         numeric(18, 2) NOT NULL,
    markup             numeric(18, 2) NOT NULL,
    sale_price         numeric(18, 2) NOT NULL,
    down_payment       numeric(18, 2) NOT NULL,
    financed_amount    numeric(18, 2) NOT NULL,
    outstanding        numeric(18, 2) NOT NULL,
    installments_count integer        NOT NULL,
    cadence            text           NOT NULL,
    currency           text           NOT NULL DEFAULT 'RUB',
    status             text           NOT NULL CHECK (status IN ('draft', 'active', 'completed', 'cancelled')),
    start_date         date           NOT NULL,
    created_at         timestamptz    NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_contracts_org ON contracts (org_id);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_contracts_client ON contracts (client_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE installments (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id uuid           NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    number      integer        NOT NULL,
    due_date    date           NOT NULL,
    amount      numeric(18, 2) NOT NULL,
    CONSTRAINT installments_contract_number_unique UNIQUE (contract_id, number)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_installments_contract ON installments (contract_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE payments (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id uuid           NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    amount      numeric(18, 2) NOT NULL,
    paid_at     timestamptz    NOT NULL,
    created_at  timestamptz    NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_payments_contract ON payments (contract_id);
-- +goose StatementEnd

-- +goose StatementBegin
-- Sadaqa registry: fixed late-payment charges that go to charity, never into the
-- contract's outstanding balance or the seller's revenue (anti-riba).
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

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS charity_entries;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS payments;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS installments;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS contracts;
-- +goose StatementEnd
