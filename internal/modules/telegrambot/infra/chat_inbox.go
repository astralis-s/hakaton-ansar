package infra

import (
	"context"

	portalapp "github.com/astralis-s/hakaton-ansar/internal/modules/portal/app"
	portaldomain "github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/telegrambot/domain"
)

// ChatInbox implements domain.ChatInbox over the portal SendMessage use-case, so
// inbound Telegram messages land in the same staff chat the managers already use.
type ChatInbox struct{ send *portalapp.SendMessage }

func NewChatInbox(send *portalapp.SendMessage) *ChatInbox {
	return &ChatInbox{send: send}
}

var _ domain.ChatInbox = (*ChatInbox)(nil)

func (c *ChatInbox) PostClientMessage(ctx context.Context, orgID, clientID, body string) error {
	_, err := c.send.Execute(ctx, portalapp.SendMessageInput{
		OrgID:      orgID,
		ClientID:   clientID,
		SenderKind: portaldomain.SenderClient,
		SenderID:   clientID,
		Body:       body,
	})
	return err
}
