// Package catalog is the composition root of the catalog bounded context.
package catalog

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/astralis-s/hakaton-ansar/internal/modules/catalog/app"
	"github.com/astralis-s/hakaton-ansar/internal/modules/catalog/domain"
	cataloghttp "github.com/astralis-s/hakaton-ansar/internal/modules/catalog/http"
	"github.com/astralis-s/hakaton-ansar/internal/modules/catalog/infra"
)

// Deps are the external dependencies of the catalog module.
type Deps struct {
	Pool *pgxpool.Pool
	Log  *slog.Logger
}

// Module is the assembled catalog module.
type Module struct {
	handler *cataloghttp.Handler
	repo    domain.ProductRepository
}

// New wires the catalog module.
func New(d Deps) *Module {
	repo := infra.NewProductRepository(d.Pool)
	handler := cataloghttp.NewHandler(cataloghttp.HandlerDeps{
		Create: app.NewCreateProduct(repo),
		Get:    app.NewGetProduct(repo),
		List:   app.NewListProducts(repo),
		Update: app.NewUpdateProduct(repo),
		Log:    d.Log,
	})
	return &Module{handler: handler, repo: repo}
}

// Products exposes the product repository for cross-context reads (financing).
func (m *Module) Products() domain.ProductRepository { return m.repo }

// RegisterRoutes mounts the catalog routes onto a JWT-protected router.
func (m *Module) RegisterRoutes(r chi.Router) {
	m.handler.RegisterRoutes(r)
}
