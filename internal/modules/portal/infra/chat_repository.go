package infra

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/infra/sqlcgen"
	"github.com/astralis-s/hakaton-ansar/internal/platform/database"
	"github.com/astralis-s/hakaton-ansar/internal/platform/pgconv"
)

// ChatRepository implements domain.ChatRepository over sqlc. AppendMessage must
// run inside a transaction (it inserts the message and bumps the conversation);
// callers provide the transaction via the context.
type ChatRepository struct{ pool *pgxpool.Pool }

func NewChatRepository(pool *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{pool: pool}
}

var _ domain.ChatRepository = (*ChatRepository)(nil)

func (r *ChatRepository) q(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(database.Querier(ctx, r.pool))
}

func (r *ChatRepository) EnsureConversation(ctx context.Context, orgID, clientID string) (domain.Conversation, error) {
	org, err := pgconv.UUID(orgID)
	if err != nil {
		return domain.Conversation{}, fmt.Errorf("invalid org id: %w", err)
	}
	cid, err := pgconv.UUID(clientID)
	if err != nil {
		return domain.Conversation{}, fmt.Errorf("invalid client id: %w", err)
	}
	newID, _ := pgconv.UUID(uuid.NewString())
	row, err := r.q(ctx).EnsureConversation(ctx, sqlcgen.EnsureConversationParams{ID: newID, OrgID: org, ClientID: cid})
	if err != nil {
		return domain.Conversation{}, fmt.Errorf("ensure conversation: %w", err)
	}
	return conversationFromRow(row), nil
}

func (r *ChatRepository) AppendMessage(ctx context.Context, m domain.Message) (domain.Message, error) {
	id, err := pgconv.UUID(m.ID())
	if err != nil {
		return domain.Message{}, fmt.Errorf("invalid message id: %w", err)
	}
	conv, err := pgconv.UUID(m.ConversationID())
	if err != nil {
		return domain.Message{}, fmt.Errorf("invalid conversation id: %w", err)
	}
	org, err := pgconv.UUID(m.OrgID())
	if err != nil {
		return domain.Message{}, fmt.Errorf("invalid org id: %w", err)
	}
	sender, err := pgconv.UUID(m.SenderID())
	if err != nil {
		return domain.Message{}, fmt.Errorf("invalid sender id: %w", err)
	}
	row, err := r.q(ctx).InsertMessage(ctx, sqlcgen.InsertMessageParams{
		ID:             id,
		ConversationID: conv,
		OrgID:          org,
		SenderKind:     m.SenderKind().String(),
		SenderID:       sender,
		Body:           m.Body(),
	})
	if err != nil {
		return domain.Message{}, fmt.Errorf("insert message: %w", err)
	}
	if err := r.q(ctx).TouchConversation(ctx, conv); err != nil {
		return domain.Message{}, fmt.Errorf("touch conversation: %w", err)
	}
	return messageFromRow(row), nil
}

func (r *ChatRepository) ListMessages(ctx context.Context, orgID, clientID string) ([]domain.Message, error) {
	org, err := pgconv.UUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid org id: %w", err)
	}
	cid, err := pgconv.UUID(clientID)
	if err != nil {
		return nil, fmt.Errorf("invalid client id: %w", err)
	}
	rows, err := r.q(ctx).ListMessagesByClient(ctx, sqlcgen.ListMessagesByClientParams{OrgID: org, ClientID: cid})
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	out := make([]domain.Message, 0, len(rows))
	for _, row := range rows {
		out = append(out, messageFromRow(row))
	}
	return out, nil
}

func (r *ChatRepository) ListConversations(ctx context.Context, orgID string) ([]domain.ConversationView, error) {
	org, err := pgconv.UUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid org id: %w", err)
	}
	rows, err := r.q(ctx).ListConversations(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	out := make([]domain.ConversationView, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.ConversationView{
			ConversationID: pgconv.StrUUID(row.ID),
			ClientID:       pgconv.StrUUID(row.ClientID),
			LastMessage:    row.LastBody,
			LastSenderKind: domain.SenderKind(row.LastSenderKind),
			LastMessageAt:  pgconv.TimeValue(row.LastMessageAt),
		})
	}
	return out, nil
}

func conversationFromRow(c sqlcgen.Conversation) domain.Conversation {
	return domain.RehydrateConversation(
		pgconv.StrUUID(c.ID),
		pgconv.StrUUID(c.OrgID),
		pgconv.StrUUID(c.ClientID),
		pgconv.TimeValue(c.CreatedAt),
		pgconv.TimeValue(c.LastMessageAt),
	)
}

func messageFromRow(m sqlcgen.Message) domain.Message {
	return domain.RehydrateMessage(
		pgconv.StrUUID(m.ID),
		pgconv.StrUUID(m.ConversationID),
		pgconv.StrUUID(m.OrgID),
		domain.SenderKind(m.SenderKind),
		pgconv.StrUUID(m.SenderID),
		m.Body,
		pgconv.TimeValue(m.CreatedAt),
	)
}
