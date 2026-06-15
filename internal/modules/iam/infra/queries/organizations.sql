-- name: CreateOrganization :one
INSERT INTO organizations (id, name, currency)
VALUES ($1, $2, $3)
RETURNING id, name, currency, created_at;

-- name: GetOrganizationByID :one
SELECT id, name, currency, created_at
FROM organizations
WHERE id = $1;

-- name: CountOrganizations :one
SELECT count(*) FROM organizations;
