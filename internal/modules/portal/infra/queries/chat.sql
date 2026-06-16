-- name: EnsureConversation :one
INSERT INTO conversations (id, org_id, client_id)
VALUES ($1, $2, $3)
ON CONFLICT (org_id, client_id) DO UPDATE SET org_id = EXCLUDED.org_id
RETURNING id, org_id, client_id, created_at, last_message_at;

-- name: InsertMessage :one
INSERT INTO messages (id, conversation_id, org_id, sender_kind, sender_id, body)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, conversation_id, org_id, sender_kind, sender_id, body, created_at;

-- name: TouchConversation :exec
UPDATE conversations SET last_message_at = now() WHERE id = $1;

-- name: ListMessagesByClient :many
SELECT m.id, m.conversation_id, m.org_id, m.sender_kind, m.sender_id, m.body, m.created_at
FROM messages m
JOIN conversations c ON c.id = m.conversation_id
WHERE c.org_id = $1 AND c.client_id = $2
ORDER BY m.created_at ASC;

-- name: ListConversations :many
SELECT c.id, c.client_id, c.last_message_at,
       COALESCE(lm.body, '')        AS last_body,
       COALESCE(lm.sender_kind, '') AS last_sender_kind
FROM conversations c
LEFT JOIN LATERAL (
    SELECT body, sender_kind
    FROM messages mm
    WHERE mm.conversation_id = c.id
    ORDER BY mm.created_at DESC
    LIMIT 1
) lm ON true
WHERE c.org_id = $1
ORDER BY c.last_message_at DESC;
