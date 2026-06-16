package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/modules/telegrambot/domain"
)

// deliveryTimeout bounds a single outbound delivery so a slow/unreachable
// Telegram API never blocks the manager's web request for long.
const deliveryTimeout = 15 * time.Second

// DeliverStaffReply pushes a manager's reply (already stored in the portal chat)
// to the client's Telegram chat, if that client is linked to one. Best-effort:
// it never returns an error to the caller — failures are logged.
type DeliverStaffReply struct {
	repo      domain.UserRepository
	messenger domain.Messenger
	log       *slog.Logger
}

func NewDeliverStaffReply(repo domain.UserRepository, messenger domain.Messenger, log *slog.Logger) *DeliverStaffReply {
	return &DeliverStaffReply{repo: repo, messenger: messenger, log: log}
}

// Execute looks up the client's Telegram chat and delivers the reply there.
func (uc *DeliverStaffReply) Execute(ctx context.Context, orgID, clientID, body string) {
	chatID, found, err := uc.repo.FindChatByClient(ctx, orgID, clientID)
	if err != nil {
		uc.log.Error("find telegram chat for client", "error", err, "client_id", clientID)
		return
	}
	if !found {
		return // this client did not come from Telegram — nothing to deliver
	}

	ctx, cancel := context.WithTimeout(ctx, deliveryTimeout)
	defer cancel()
	if err := uc.messenger.Send(ctx, chatID, body); err != nil {
		uc.log.Error("deliver staff reply to telegram", "error", err, "chat_id", chatID)
	}
}
