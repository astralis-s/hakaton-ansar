-- +goose Up
-- +goose StatementBegin
-- Baseline migration. pgcrypto provides gen_random_uuid(), used by later
-- module migrations (iam, catalog, crm, financing, scheduling).
CREATE EXTENSION IF NOT EXISTS pgcrypto;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP EXTENSION IF EXISTS pgcrypto;
-- +goose StatementEnd
