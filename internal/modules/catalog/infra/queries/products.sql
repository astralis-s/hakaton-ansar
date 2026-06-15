-- name: CreateProduct :one
INSERT INTO products (id, org_id, name, category, cost_price, halal_status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, org_id, name, category, cost_price, halal_status, created_at;

-- name: GetProductByID :one
SELECT id, org_id, name, category, cost_price, halal_status, created_at
FROM products
WHERE id = $1 AND org_id = $2;

-- name: ListProductsByOrg :many
SELECT id, org_id, name, category, cost_price, halal_status, created_at
FROM products
WHERE org_id = $1
ORDER BY created_at DESC;

-- name: UpdateProduct :one
UPDATE products
SET name = $3, category = $4, cost_price = $5, halal_status = $6
WHERE id = $1 AND org_id = $2
RETURNING id, org_id, name, category, cost_price, halal_status, created_at;
