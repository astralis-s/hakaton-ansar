// Package iam is the composition root of the IAM bounded context: it wires
// repositories → use-cases → HTTP handler and exposes the auth middleware used
// by both HTTP surfaces.
package iam

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/astralis-s/hakaton-ansar/internal/modules/iam/app"
	"github.com/astralis-s/hakaton-ansar/internal/modules/iam/domain"
	iamhttp "github.com/astralis-s/hakaton-ansar/internal/modules/iam/http"
	"github.com/astralis-s/hakaton-ansar/internal/modules/iam/infra"
)

// Deps are the external dependencies of the iam module.
type Deps struct {
	Pool      *pgxpool.Pool
	Tx        domain.TxManager
	Log       *slog.Logger
	JWTSecret string
	JWTTTL    time.Duration
}

// Module is the assembled iam module.
type Module struct {
	handler  *iamhttp.Handler
	jwtMW    func(http.Handler) http.Handler
	apiKeyMW func(http.Handler) http.Handler
	ownerMW  func(http.Handler) http.Handler
	orgs     domain.OrganizationRepository
}

// New wires the iam module.
func New(d Deps) *Module {
	orgRepo := infra.NewOrganizationRepository(d.Pool)
	userRepo := infra.NewUserRepository(d.Pool)
	keyRepo := infra.NewApiKeyRepository(d.Pool)
	hasher := infra.NewBcryptHasher()
	tokens := infra.NewJWTService(d.JWTSecret, d.JWTTTL)

	authKeyUC := app.NewAuthenticateApiKey(keyRepo)
	registerUC := app.NewRegisterOrganization(orgRepo, userRepo, hasher, d.Tx)

	handler := iamhttp.NewHandler(iamhttp.HandlerDeps{
		Setup:      app.NewSetupOrganization(orgRepo, registerUC),
		Register:   registerUC,
		Login:      app.NewLogin(userRepo, hasher, tokens),
		CreateUser: app.NewCreateUser(userRepo, hasher),
		ListUsers:  app.NewListUsers(userRepo),
		GetUser:    app.NewGetUser(userRepo),
		CreateKey:  app.NewCreateApiKey(keyRepo),
		ListKeys:   app.NewListApiKeys(keyRepo),
		RevokeKey:  app.NewRevokeApiKey(keyRepo),
		Log:        d.Log,
	})

	return &Module{
		handler:  handler,
		jwtMW:    iamhttp.JWTAuth(tokens, d.Log),
		apiKeyMW: iamhttp.APIKeyAuth(authKeyUC, d.Log),
		ownerMW:  iamhttp.RequireOwner(d.Log),
		orgs:     orgRepo,
	}
}

// Organizations exposes the organization repository for cross-context reads
// (e.g. resolving the seller name on contract documents).
func (m *Module) Organizations() domain.OrganizationRepository { return m.orgs }

// JWTMiddleware protects the internal /api/app surface (Authorization: Bearer).
func (m *Module) JWTMiddleware() func(http.Handler) http.Handler { return m.jwtMW }

// APIKeyMiddleware protects the public /api/v1 surface (X-API-Key).
func (m *Module) APIKeyMiddleware() func(http.Handler) http.Handler { return m.apiKeyMW }

// OwnerMiddleware rejects non-owner principals (403). Other modules use it to
// gate owner-only actions (e.g. cancel contract).
func (m *Module) OwnerMiddleware() func(http.Handler) http.Handler { return m.ownerMW }

// RegisterPublicAppRoutes mounts unauthenticated /api/app routes.
func (m *Module) RegisterPublicAppRoutes(r chi.Router) {
	r.Post("/auth/login", m.handler.Login)
	r.Post("/auth/register", m.handler.Register) // multi-tenant sign-up
	r.Post("/setup", m.handler.Setup)            // first-run alias
}

// RegisterProtectedAppRoutes mounts JWT-protected /api/app routes. The caller
// must have applied JWTMiddleware to r.
func (m *Module) RegisterProtectedAppRoutes(r chi.Router) {
	r.Get("/auth/me", m.handler.Me)

	r.Route("/users", func(ur chi.Router) {
		ur.Use(m.ownerMW) // user management is owner-only (settings/users)
		ur.Get("/", m.handler.ListUsers)
		ur.Post("/", m.handler.CreateUser)
	})

	r.Route("/api-keys", func(kr chi.Router) {
		kr.Use(m.ownerMW) // API keys are owner-only (developers/api-keys)
		kr.Get("/", m.handler.ListApiKeys)
		kr.Post("/", m.handler.CreateApiKey)
		kr.Delete("/{id}", m.handler.RevokeApiKey)
	})
}
