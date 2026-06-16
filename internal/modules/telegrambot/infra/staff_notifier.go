package infra

import (
	"context"

	portaldomain "github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/telegrambot/app"
)

// StaffNotifier implements the portal StaffReplyNotifier port: whenever a manager
// sends a reply in the staff chat, it is delivered to the client's Telegram chat
// (if any). This is the outbound half of the bridge.
type StaffNotifier struct{ deliver *app.DeliverStaffReply }

func NewStaffNotifier(deliver *app.DeliverStaffReply) *StaffNotifier {
	return &StaffNotifier{deliver: deliver}
}

var _ portaldomain.StaffReplyNotifier = (*StaffNotifier)(nil)

func (n *StaffNotifier) StaffReplied(ctx context.Context, orgID, clientID, body string) {
	n.deliver.Execute(ctx, orgID, clientID, body)
}
