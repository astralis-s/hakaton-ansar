-- name: CreateContractRequest :one
INSERT INTO contract_requests (id, org_id, client_id, product_id, desired_installments, desired_down_payment, note, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, org_id, client_id, product_id, desired_installments, desired_down_payment, note, status, contract_id, created_at, decided_at;

-- name: GetContractRequestByID :one
SELECT id, org_id, client_id, product_id, desired_installments, desired_down_payment, note, status, contract_id, created_at, decided_at
FROM contract_requests
WHERE id = $1 AND org_id = $2;

-- name: ListContractRequestsByOrg :many
SELECT id, org_id, client_id, product_id, desired_installments, desired_down_payment, note, status, contract_id, created_at, decided_at
FROM contract_requests
WHERE org_id = $1
ORDER BY created_at DESC;

-- name: ListContractRequestsByClient :many
SELECT id, org_id, client_id, product_id, desired_installments, desired_down_payment, note, status, contract_id, created_at, decided_at
FROM contract_requests
WHERE org_id = $1 AND client_id = $2
ORDER BY created_at DESC;

-- name: UpdateContractRequest :exec
UPDATE contract_requests
SET status = $3, contract_id = $4, decided_at = $5
WHERE id = $1 AND org_id = $2;
