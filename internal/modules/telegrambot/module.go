// Package telegrambot is the composition root of the telegram-bot bounded
// context. It bridges Telegram customers to the existing staff chat: inbound
// messages are relayed into the portal inbox (under the client's full name) and
// manager replies are delivered back to Telegram. The bot is optional — it only
// runs when a token is configured.
package telegrambot

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	portaldomain "github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/telegrambot/app"
	"github.com/astralis-s/hakaton-ansar/internal/modules/telegrambot/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/telegrambot/infra"
	"github.com/astralis-s/hakaton-ansar/internal/modules/telegrambot/infra/tgapi"
)

// Deps are the external dependencies of the telegrambot module. Clients/Inbox
// are cross-context adapters (wired from crm/portal in main).
type Deps struct {
	Pool            *pgxpool.Pool
	Tx              domain.TxManager
	Log             *slog.Logger
	BotToken        string
	PollTimeout     time.Duration
	ConfiguredOrgID string // optional; empty → resolve the single organization
	Clients         domain.ClientDirectory
	Inbox           domain.ChatInbox
}

// Module is the assembled telegrambot module.
type Module struct {
	client          *tgapi.Client
	poller          *tgapi.Poller
	repo            *infra.UserRepository
	messenger       *infra.Messenger
	notifier        *infra.StaffNotifier
	linkProvider    *infra.LinkProvider
	clients         domain.ClientDirectory
	inbox           domain.ChatInbox
	tx              domain.TxManager
	configuredOrgID string
	log             *slog.Logger
}

// New wires the telegrambot module.
func New(d Deps) *Module {
	client := tgapi.New(d.BotToken)
	repo := infra.NewUserRepository(d.Pool)
	messenger := infra.NewMessenger(client)
	deliver := app.NewDeliverStaffReply(repo, messenger, d.Log)
	linkProvider := infra.NewLinkProvider(app.NewManagerLink(infra.NewBotIdentity(client)))

	return &Module{
		client:          client,
		poller:          tgapi.NewPoller(client, d.Log, int(d.PollTimeout.Seconds())),
		repo:            repo,
		messenger:       messenger,
		notifier:        infra.NewStaffNotifier(deliver),
		linkProvider:    linkProvider,
		clients:         d.Clients,
		inbox:           d.Inbox,
		tx:              d.Tx,
		configuredOrgID: d.ConfiguredOrgID,
		log:             d.Log,
	}
}

// Notifier returns the portal StaffReplyNotifier that delivers manager replies to
// Telegram. Wire it into the portal module so staff replies reach the bot.
func (m *Module) Notifier() portaldomain.StaffReplyNotifier { return m.notifier }

// LinkProvider returns the portal TelegramLinkProvider that builds each manager's
// personal bot deep link for the staff chat page.
func (m *Module) LinkProvider() portaldomain.TelegramLinkProvider { return m.linkProvider }

// Run validates the token, resolves the organization and long-polls Telegram
// until ctx is cancelled. Intended to be launched in its own goroutine.
func (m *Module) Run(ctx context.Context) {
	me, err := m.client.GetMe(ctx)
	if err != nil {
		m.log.Error("telegram bot: invalid token (getMe failed); bot disabled", "error", err)
		return
	}

	orgID := m.configuredOrgID
	if orgID == "" {
		resolved, err := m.repo.DefaultOrgID(ctx)
		if err != nil {
			m.log.Error("telegram bot: resolve organization", "error", err)
			return
		}
		orgID = resolved
	}
	if orgID == "" {
		m.log.Warn("telegram bot: no organization found; bot will not start (run setup or seed first)")
		return
	}

	process := app.NewProcessUpdate(m.repo, m.clients, m.inbox, m.messenger, m.tx, orgID, m.log)
	m.log.Info("telegram bot: polling started", "username", me.Username, "org_id", orgID)

	m.poller.Run(ctx, func(ctx context.Context, u tgapi.Update) {
		if in, ok := toIncoming(u); ok {
			process.HandleMessage(ctx, in)
		}
	})
}

// toIncoming maps a Telegram update onto the transport-agnostic IncomingMessage,
// keeping Telegram types out of the application layer.
func toIncoming(u tgapi.Update) (app.IncomingMessage, bool) {
	if u.Message == nil || u.Message.Chat.Type != "private" {
		return app.IncomingMessage{}, false
	}
	in := app.IncomingMessage{ChatID: u.Message.Chat.ID, Text: u.Message.Text}
	if u.Message.From != nil {
		in.Username = u.Message.From.Username
	}
	if u.Message.Contact != nil {
		in.Contact = u.Message.Contact.PhoneNumber
	}
	return in, true
}
