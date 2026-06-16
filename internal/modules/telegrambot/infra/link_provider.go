package infra

import (
	"context"

	portaldomain "github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/telegrambot/app"
)

// LinkProvider implements the portal TelegramLinkProvider port, so the staff chat
// page can show each manager their personal bot deep link.
type LinkProvider struct{ link *app.ManagerLink }

func NewLinkProvider(link *app.ManagerLink) *LinkProvider {
	return &LinkProvider{link: link}
}

var _ portaldomain.TelegramLinkProvider = (*LinkProvider)(nil)

func (p *LinkProvider) ManagerLink(ctx context.Context, _ /* orgID */, managerID string) (string, bool, error) {
	url, err := p.link.Execute(ctx, managerID)
	if err != nil {
		return "", false, err
	}
	if url == "" {
		return "", false, nil
	}
	return url, true, nil
}
