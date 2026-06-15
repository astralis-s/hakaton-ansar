-- name: CreateCharityEntry :one
INSERT INTO charity_entries (id, org_id, contract_id, client_id, amount, status, note, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, org_id, contract_id, client_id, amount, status, note, created_by, created_at;

-- name: ListCharityByOrg :many
SELECT id, org_id, contract_id, client_id, amount, status, note, created_by, created_at
FROM charity_entries
WHERE org_id = $1
ORDER BY created_at DESC;
