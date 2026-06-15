-- name: CreateApiKey :one
INSERT INTO api_keys (id, org_id, name, prefix, key_hash)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, org_id, name, prefix, key_hash, created_at, revoked_at;

-- name: GetApiKeyByHash :one
SELECT id, org_id, name, prefix, key_hash, created_at, revoked_at
FROM api_keys
WHERE key_hash = $1;

-- name: ListApiKeysByOrg :many
SELECT id, org_id, name, prefix, key_hash, created_at, revoked_at
FROM api_keys
WHERE org_id = $1
ORDER BY created_at DESC;

-- name: RevokeApiKey :one
UPDATE api_keys
SET revoked_at = now()
WHERE id = $1 AND org_id = $2 AND revoked_at IS NULL
RETURNING id, org_id, name, prefix, key_hash, created_at, revoked_at;
