package infra

import (
	"context"

	"github.com/astralis-s/hakaton-ansar/internal/modules/telegrambot/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/telegrambot/infra/tgapi"
)

// Messenger implements domain.Messenger over the Telegram Bot API client.
type Messenger struct{ client *tgapi.Client }

func NewMessenger(client *tgapi.Client) *Messenger {
	return &Messenger{client: client}
}

var _ domain.Messenger = (*Messenger)(nil)

// Send delivers plain text and clears any custom keyboard left over from the
// registration flow.
func (m *Messenger) Send(ctx context.Context, chatID int64, text string) error {
	return m.client.SendMessageWithMarkup(ctx, chatID, text, tgapi.ReplyKeyboardRemove{RemoveKeyboard: true})
}

// AskForContact delivers text together with the "share phone" keyboard button.
func (m *Messenger) AskForContact(ctx context.Context, chatID int64, text string) error {
	return m.client.SendMessageWithMarkup(ctx, chatID, text, tgapi.ContactKeyboard())
}
