package infra

import (
	"context"
	"encoding/base64"

	qrcode "github.com/skip2/go-qrcode"

	portaldomain "github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/telegrambot/app"
)

// LinkProvider implements the portal TelegramLinkProvider port, so the staff chat
// page and the owner's settings can show each manager their personal bot deep
// link together with a scannable QR code of it.
type LinkProvider struct{ link *app.ManagerLink }

func NewLinkProvider(link *app.ManagerLink) *LinkProvider {
	return &LinkProvider{link: link}
}

var _ portaldomain.TelegramLinkProvider = (*LinkProvider)(nil)

func (p *LinkProvider) ManagerLink(ctx context.Context, _ /* orgID */, managerID string) (string, string, bool, error) {
	url, err := p.link.Execute(ctx, managerID)
	if err != nil {
		return "", "", false, err
	}
	if url == "" {
		return "", "", false, nil
	}
	qr, err := qrDataURI(url)
	if err != nil {
		return "", "", false, err
	}
	return url, qr, true, nil
}

// qrDataURI renders content as a PNG QR code, base64-encoded into a data URI so it
// can be embedded directly in an <img src>.
func qrDataURI(content string) (string, error) {
	png, err := qrcode.Encode(content, qrcode.Medium, 320)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), nil
}
