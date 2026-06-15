-- +goose Up
-- +goose StatementBegin
CREATE TABLE products (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       uuid          NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name         text          NOT NULL,
    category     text          NOT NULL DEFAULT '',
    cost_price   numeric(18, 2) NOT NULL CHECK (cost_price > 0),
    halal_status text          NOT NULL CHECK (halal_status IN ('halal', 'haram', 'doubtful')),
    created_at   timestamptz   NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_products_org ON products (org_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS products;
-- +goose StatementEnd
