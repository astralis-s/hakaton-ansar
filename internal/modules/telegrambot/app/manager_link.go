package app

import (
	"context"
	"strings"

	"github.com/astralis-s/hakaton-ansar/internal/modules/telegrambot/domain"
)

// ManagerLink builds a manager's personal Telegram deep link. Opening it and
// pressing Start sends the bot "/start <payload>", where the payload carries the
// manager id — so each manager has their own shareable link that launches the
// bot and starts the registration flow.
type ManagerLink struct {
	bot domain.BotIdentity
}

func NewManagerLink(bot domain.BotIdentity) *ManagerLink {
	return &ManagerLink{bot: bot}
}

// Execute returns the deep link for a manager, or "" when the bot username is
// not available yet (e.g. token misconfigured).
func (uc *ManagerLink) Execute(ctx context.Context, managerID string) (string, error) {
	username, err := uc.bot.Username(ctx)
	if err != nil {
		return "", err
	}
	if username == "" || managerID == "" {
		return "", nil
	}
	return "https://t.me/" + username + "?start=" + EncodeManagerPayload(managerID), nil
}

// EncodeManagerPayload encodes a manager id into a Telegram start payload (the
// allowed charset is A-Za-z0-9_-, so we drop the UUID hyphens).
func EncodeManagerPayload(managerID string) string {
	return strings.ReplaceAll(managerID, "-", "")
}
