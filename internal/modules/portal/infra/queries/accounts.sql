-- name: UpsertPortalAccount :exec
INSERT INTO client_portal_accounts (client_id, org_id, email, password_hash)
VALUES ($1, $2, $3, $4)
ON CONFLICT (client_id) DO UPDATE
    SET email = EXCLUDED.email, password_hash = EXCLUDED.password_hash;

-- name: GetPortalAccountByEmail :one
SELECT client_id, org_id, email, password_hash, created_at
FROM client_portal_accounts
WHERE email = $1;

-- name: GetPortalAccountByClientID :one
SELECT client_id, org_id, email, password_hash, created_at
FROM client_portal_accounts
WHERE org_id = $1 AND client_id = $2;
