-- name: CreateContract :exec
INSERT INTO contracts (
    id, org_id, client_id, product_id,
    cost_price, markup, sale_price, down_payment, financed_amount, outstanding,
    installments_count, cadence, currency, status, start_date
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15
);

-- name: CreateInstallment :exec
INSERT INTO installments (id, contract_id, number, due_date, amount)
VALUES ($1, $2, $3, $4, $5);

-- name: CreatePayment :exec
INSERT INTO payments (id, contract_id, amount, paid_at)
VALUES ($1, $2, $3, $4);

-- name: UpdateContractState :exec
UPDATE contracts
SET outstanding = $3, status = $4
WHERE id = $1 AND org_id = $2;

-- name: GetContractByID :one
SELECT id, org_id, client_id, product_id,
       cost_price, markup, sale_price, down_payment, financed_amount, outstanding,
       installments_count, cadence, currency, status, start_date, created_at
FROM contracts
WHERE id = $1 AND org_id = $2;

-- name: ListContractsByOrg :many
SELECT id, org_id, client_id, product_id,
       cost_price, markup, sale_price, down_payment, financed_amount, outstanding,
       installments_count, cadence, currency, status, start_date, created_at
FROM contracts
WHERE org_id = $1
ORDER BY created_at DESC;

-- name: ListInstallmentsByContract :many
SELECT id, contract_id, number, due_date, amount
FROM installments
WHERE contract_id = $1
ORDER BY number ASC;

-- name: ListPaymentsByContract :many
SELECT id, contract_id, amount, paid_at, created_at
FROM payments
WHERE contract_id = $1
ORDER BY paid_at ASC;
