-- name: CreateClient :one
INSERT INTO clients (id, org_id, full_name, phone, document)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, org_id, full_name, phone, document, created_at;

-- name: GetClientByID :one
SELECT id, org_id, full_name, phone, document, created_at
FROM clients
WHERE id = $1 AND org_id = $2;

-- name: ListClientsByOrg :many
SELECT id, org_id, full_name, phone, document, created_at
FROM clients
WHERE org_id = $1
ORDER BY created_at DESC;

-- name: UpdateClient :one
UPDATE clients
SET full_name = $3, phone = $4, document = $5
WHERE id = $1 AND org_id = $2
RETURNING id, org_id, full_name, phone, document, created_at;
