-- name: CreateReminder :one
INSERT INTO reminders (
    id, org_id, type, client_id, contract_id, note,
    desired_at, duration_minutes, scheduled_at, was_shifted, reason, status
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10, $11, $12
)
RETURNING id, org_id, type, client_id, contract_id, note,
          desired_at, duration_minutes, scheduled_at, was_shifted, reason, status, completed_at, cancelled_at, created_at;

-- name: GetReminderByID :one
SELECT id, org_id, type, client_id, contract_id, note,
       desired_at, duration_minutes, scheduled_at, was_shifted, reason, status, completed_at, cancelled_at, created_at
FROM reminders
WHERE org_id = $1 AND id = $2;

-- name: ListRemindersByOrg :many
SELECT id, org_id, type, client_id, contract_id, note,
       desired_at, duration_minutes, scheduled_at, was_shifted, reason, status, completed_at, cancelled_at, created_at
FROM reminders
WHERE org_id = $1
ORDER BY scheduled_at ASC;

-- name: UpdateReminder :one
UPDATE reminders
SET type = $3,
    client_id = $4,
    contract_id = $5,
    note = $6,
    desired_at = $7,
    duration_minutes = $8,
    scheduled_at = $9,
    was_shifted = $10,
    reason = $11,
    status = $12,
    completed_at = $13,
    cancelled_at = $14
WHERE org_id = $1 AND id = $2
RETURNING id, org_id, type, client_id, contract_id, note,
          desired_at, duration_minutes, scheduled_at, was_shifted, reason, status, completed_at, cancelled_at, created_at;
