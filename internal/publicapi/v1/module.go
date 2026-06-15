package v1

import (
	"log/slog"

	"github.com/go-chi/chi/v5"

	financingapp "github.com/astralis-s/hakaton-ansar/internal/modules/financing/app"
)

// Deps are the dependencies of the public API: the same financing use-cases the
// internal API uses (no logic duplication).
type Deps struct {
	CreateContract *financingapp.CreateContract
	GetContract    *financingapp.GetContract
	Log            *slog.Logger
}

// Module is the assembled public API.
type Module struct {
	handler *Handler
}

// New wires the public API module.
func New(d Deps) *Module {
	return &Module{handler: NewHandler(d.CreateContract, d.GetContract, d.Log)}
}

// RegisterRoutes mounts the public API routes (caller applies X-API-Key mw).
func (m *Module) RegisterRoutes(r chi.Router) {
	m.handler.RegisterRoutes(r)
}
