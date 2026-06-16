package infra

import (
	"context"
	"sync"

	"github.com/astralis-s/hakaton-ansar/internal/modules/telegrambot/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/telegrambot/infra/tgapi"
)

// BotIdentity implements domain.BotIdentity, caching the bot @username from getMe
// after the first successful lookup.
type BotIdentity struct {
	client *tgapi.Client
	mu     sync.Mutex
	cached string
}

func NewBotIdentity(client *tgapi.Client) *BotIdentity {
	return &BotIdentity{client: client}
}

var _ domain.BotIdentity = (*BotIdentity)(nil)

func (b *BotIdentity) Username(ctx context.Context) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.cached != "" {
		return b.cached, nil
	}
	me, err := b.client.GetMe(ctx)
	if err != nil {
		return "", err
	}
	b.cached = me.Username
	return b.cached, nil
}
