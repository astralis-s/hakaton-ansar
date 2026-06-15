-- name: CreateReminder :one
INSERT INTO reminders (
    id, org_id, type, client_id, contract_id, note,
    desired_at, duration_minutes, scheduled_at, was_shifted, reason
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10, $11
)
RETURNING id, org_id, type, client_id, contract_id, note,
          desired_at, duration_minutes, scheduled_at, was_shifted, reason, created_at;

-- name: ListRemindersByOrg :many
SELECT id, org_id, type, client_id, contract_id, note,
       desired_at, duration_minutes, scheduled_at, was_shifted, reason, created_at
FROM reminders
WHERE org_id = $1
ORDER BY scheduled_at ASC;
