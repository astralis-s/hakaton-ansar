// Package crm is the composition root of the crm bounded context.
package crm

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/astralis-s/hakaton-ansar/internal/modules/crm/app"
	crmhttp "github.com/astralis-s/hakaton-ansar/internal/modules/crm/http"
	"github.com/astralis-s/hakaton-ansar/internal/modules/crm/infra"
)

// Deps are the external dependencies of the crm module.
type Deps struct {
	Pool *pgxpool.Pool
	Log  *slog.Logger
}

// Module is the assembled crm module.
type Module struct {
	handler *crmhttp.Handler
}

// New wires the crm module.
func New(d Deps) *Module {
	repo := infra.NewClientRepository(d.Pool)
	handler := crmhttp.NewHandler(crmhttp.HandlerDeps{
		Create: app.NewCreateClient(repo),
		Get:    app.NewGetClient(repo),
		List:   app.NewListClients(repo),
		Update: app.NewUpdateClient(repo),
		Log:    d.Log,
	})
	return &Module{handler: handler}
}

// RegisterRoutes mounts the crm routes onto a JWT-protected router.
func (m *Module) RegisterRoutes(r chi.Router) {
	m.handler.RegisterRoutes(r)
}
