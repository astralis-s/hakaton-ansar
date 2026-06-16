-- +goose Up
-- +goose StatementBegin
ALTER TABLE products ADD COLUMN stock integer NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose StatementBegin
-- Stock movements (товарооборот): every change to a product's balance is logged.
CREATE TABLE stock_movements (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        uuid        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    product_id    uuid        NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    delta         integer     NOT NULL,
    reason        text        NOT NULL CHECK (reason IN ('receipt', 'sale', 'adjustment', 'writeoff')),
    note          text        NOT NULL DEFAULT '',
    balance_after integer     NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_stock_movements_org ON stock_movements (org_id);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_stock_movements_product ON stock_movements (product_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS stock_movements;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE products DROP COLUMN IF EXISTS stock;
-- +goose StatementEnd
