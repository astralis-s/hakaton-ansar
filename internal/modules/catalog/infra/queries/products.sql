-- name: CreateProduct :one
INSERT INTO products (id, org_id, name, category, cost_price, halal_status, stock)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, org_id, name, category, cost_price, halal_status, created_at, stock;

-- name: GetProductByID :one
SELECT id, org_id, name, category, cost_price, halal_status, created_at, stock
FROM products
WHERE id = $1 AND org_id = $2;

-- name: GetProductForUpdate :one
SELECT id, org_id, name, category, cost_price, halal_status, created_at, stock
FROM products
WHERE id = $1 AND org_id = $2
FOR UPDATE;

-- name: ListProductsByOrg :many
SELECT id, org_id, name, category, cost_price, halal_status, created_at, stock
FROM products
WHERE org_id = $1
ORDER BY created_at DESC;

-- name: UpdateProduct :one
UPDATE products
SET name = $3, category = $4, cost_price = $5, halal_status = $6
WHERE id = $1 AND org_id = $2
RETURNING id, org_id, name, category, cost_price, halal_status, created_at, stock;

-- name: SetProductStock :one
UPDATE products
SET stock = $3
WHERE id = $1 AND org_id = $2
RETURNING id, org_id, name, category, cost_price, halal_status, created_at, stock;

-- name: CreateStockMovement :one
INSERT INTO stock_movements (id, org_id, product_id, delta, reason, note, balance_after)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, org_id, product_id, delta, reason, note, balance_after, created_at;

-- name: ListStockMovementsByOrg :many
SELECT id, org_id, product_id, delta, reason, note, balance_after, created_at
FROM stock_movements
WHERE org_id = $1
ORDER BY created_at DESC;
