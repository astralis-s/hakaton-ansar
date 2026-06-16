// Package portal is the composition root of the client-portal bounded context:
// client login + the client↔staff chat.
package portal

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/app"
	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
	portalhttp "github.com/astralis-s/hakaton-ansar/internal/modules/portal/http"
	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/infra"
)

// Deps are the external dependencies of the portal module. Clients/Contracts are
// cross-context readers (wired from crm/financing in main).
type Deps struct {
	Pool      *pgxpool.Pool
	Tx        domain.TxManager
	Log       *slog.Logger
	JWTSecret string
	JWTTTL    time.Duration
	Clients   domain.ClientReader
	Contracts domain.ContractReader
	Catalog   domain.CatalogReader
	Requests  domain.RequestService
	Doc       domain.ContractDocBuilder
}

// Module is the assembled portal module.
type Module struct {
	handler  *portalhttp.Handler
	clientMW func(http.Handler) http.Handler
	send     *app.SendMessage
}

// New wires the portal module.
func New(d Deps) *Module {
	accounts := infra.NewAccountRepository(d.Pool)
	chat := infra.NewChatRepository(d.Pool)
	hasher := infra.NewBcryptHasher()
	tokens := infra.NewClientJWTService(d.JWTSecret, d.JWTTTL)

	send := app.NewSendMessage(chat, d.Tx)
	handler := portalhttp.NewHandler(portalhttp.HandlerDeps{
		Provision: app.NewProvisionAccess(accounts, d.Clients, hasher),
		GetAccess: app.NewGetAccess(accounts),
		Login:     app.NewLoginClient(accounts, hasher, tokens),
		Send:      send,
		ListConv:  app.NewListConversations(chat, d.Clients),
		Thread:    app.NewGetThread(chat),
		Profile:   app.NewGetClientProfile(d.Clients),
		Contracts: app.NewGetClientContracts(d.Contracts),
		Contract:  app.NewGetClientContract(d.Contracts),
		Browse:    app.NewBrowseProducts(d.Catalog),
		SubmitReq: app.NewSubmitRequest(d.Requests),
		MyReqs:    app.NewListMyRequests(d.Requests),
		Doc:       app.NewGetContractDoc(d.Doc),
		Log:       d.Log,
	})
	return &Module{handler: handler, clientMW: portalhttp.ClientAuth(tokens, d.Log), send: send}
}

// ClientMiddleware validates the client portal JWT for the /api/portal surface.
func (m *Module) ClientMiddleware() func(http.Handler) http.Handler { return m.clientMW }

// SendMessageUseCase exposes the chat SendMessage use-case so other contexts
// (e.g. the Telegram bot) can post inbound client messages into the staff chat.
func (m *Module) SendMessageUseCase() *app.SendMessage { return m.send }

// SetStaffReplyNotifier wires an external delivery channel for staff replies
// (e.g. the Telegram bot). Called from main after both modules are built.
func (m *Module) SetStaffReplyNotifier(n domain.StaffReplyNotifier) { m.send.SetNotifier(n) }

// SetTelegramLinkProvider wires the manager-deep-link provider shown on the staff
// chat page. Called from main when the Telegram bot is enabled.
func (m *Module) SetTelegramLinkProvider(p domain.TelegramLinkProvider) { m.handler.SetTelegramLink(p) }

// RegisterStaffRoutes mounts staff chat + portal-access routes (JWT /api/app).
func (m *Module) RegisterStaffRoutes(r chi.Router) { m.handler.RegisterStaffRoutes(r) }

// RegisterPublicPortalRoutes mounts the unauthenticated client login route.
func (m *Module) RegisterPublicPortalRoutes(r chi.Router) { m.handler.RegisterPublicPortalRoutes(r) }

// RegisterProtectedPortalRoutes mounts the client-JWT-protected portal routes.
func (m *Module) RegisterProtectedPortalRoutes(r chi.Router) {
	m.handler.RegisterProtectedPortalRoutes(r)
}
