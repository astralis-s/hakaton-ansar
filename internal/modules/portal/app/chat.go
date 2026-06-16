package app

import (
	"context"

	"github.com/google/uuid"

	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
)

// SendMessage appends a message to a client's conversation (creating the
// conversation on first use), as the client or as staff. The ensure-conversation
// and append run in one transaction.
type SendMessage struct {
	chat domain.ChatRepository
	tx   domain.TxManager
}

func NewSendMessage(chat domain.ChatRepository, tx domain.TxManager) *SendMessage {
	return &SendMessage{chat: chat, tx: tx}
}

type SendMessageInput struct {
	OrgID      string
	ClientID   string
	SenderKind domain.SenderKind
	SenderID   string
	Body       string
}

func (uc *SendMessage) Execute(ctx context.Context, in SendMessageInput) (domain.Message, error) {
	var msg domain.Message
	err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		conv, err := uc.chat.EnsureConversation(ctx, in.OrgID, in.ClientID)
		if err != nil {
			return err
		}
		m, err := domain.NewMessage(uuid.NewString(), conv.ID(), in.OrgID, in.SenderKind, in.SenderID, in.Body)
		if err != nil {
			return err
		}
		stored, err := uc.chat.AppendMessage(ctx, m)
		if err != nil {
			return err
		}
		msg = stored
		return nil
	})
	return msg, err
}

// ListConversations returns the staff inbox (conversations + last-message
// preview) with each client's display name resolved.
type ListConversations struct {
	chat    domain.ChatRepository
	clients domain.ClientReader
}

func NewListConversations(chat domain.ChatRepository, clients domain.ClientReader) *ListConversations {
	return &ListConversations{chat: chat, clients: clients}
}

func (uc *ListConversations) Execute(ctx context.Context, orgID string) ([]domain.ConversationView, map[string]string, error) {
	convs, err := uc.chat.ListConversations(ctx, orgID)
	if err != nil {
		return nil, nil, err
	}
	ids := make([]string, 0, len(convs))
	for _, c := range convs {
		ids = append(ids, c.ClientID)
	}
	names, err := uc.clients.Names(ctx, orgID, ids)
	if err != nil {
		return nil, nil, err
	}
	return convs, names, nil
}

// GetThread returns a client's full message thread (chronological).
type GetThread struct {
	chat domain.ChatRepository
}

func NewGetThread(chat domain.ChatRepository) *GetThread {
	return &GetThread{chat: chat}
}

func (uc *GetThread) Execute(ctx context.Context, orgID, clientID string) ([]domain.Message, error) {
	return uc.chat.ListMessages(ctx, orgID, clientID)
}
