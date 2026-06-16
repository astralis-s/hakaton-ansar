-- name: GetTelegramUserByChatID :one
SELECT chat_id, org_id, client_id, username, full_name, phone, state, created_at, updated_at
FROM telegram_users
WHERE chat_id = $1;

-- name: UpsertTelegramUser :one
INSERT INTO telegram_users (chat_id, org_id, client_id, username, full_name, phone, state)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (chat_id) DO UPDATE SET
    org_id = EXCLUDED.org_id,
    client_id = EXCLUDED.client_id,
    username = EXCLUDED.username,
    full_name = EXCLUDED.full_name,
    phone = EXCLUDED.phone,
    state = EXCLUDED.state,
    updated_at = now()
RETURNING chat_id, org_id, client_id, username, full_name, phone, state, created_at, updated_at;

-- name: GetChatIDByClient :one
SELECT chat_id
FROM telegram_users
WHERE org_id = $1 AND client_id = $2 AND state = 'registered';

-- name: GetDefaultOrgID :one
SELECT id
FROM organizations
ORDER BY created_at
LIMIT 1;
