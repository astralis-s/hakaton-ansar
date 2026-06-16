-- name: CreateExpense :one
INSERT INTO expenses (id, org_id, category, amount, note, spent_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, org_id, category, amount, note, spent_at, created_at;

-- name: ListExpensesByOrg :many
SELECT id, org_id, category, amount, note, spent_at, created_at
FROM expenses
WHERE org_id = $1
ORDER BY spent_at DESC, created_at DESC;

-- name: DeleteExpense :execrows
DELETE FROM expenses
WHERE id = $1 AND org_id = $2;
