-- name: CreateUser :one
INSERT INTO users (id, org_id, full_name, email, password_hash, role)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, org_id, full_name, email, password_hash, role, created_at;

-- name: GetUserByEmail :one
SELECT id, org_id, full_name, email, password_hash, role, created_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, org_id, full_name, email, password_hash, role, created_at
FROM users
WHERE id = $1;

-- name: ListUsersByOrg :many
SELECT id, org_id, full_name, email, password_hash, role, created_at
FROM users
WHERE org_id = $1
ORDER BY created_at ASC;

-- name: CountUsersByEmail :one
SELECT count(*) FROM users WHERE email = $1;
