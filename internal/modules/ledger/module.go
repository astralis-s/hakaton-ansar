// Package ledger is the composition root of the ledger bounded context
// (income/expense accounting — учёт доходов и расходов).
package ledger

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/astralis-s/hakaton-ansar/internal/modules/ledger/app"
	"github.com/astralis-s/hakaton-ansar/internal/modules/ledger/domain"
	ledgerhttp "github.com/astralis-s/hakaton-ansar/internal/modules/ledger/http"
	"github.com/astralis-s/hakaton-ansar/internal/modules/ledger/infra"
)

// Deps are the external dependencies of the ledger module. Sales is the
// cross-context reader over financing (wired in main).
type Deps struct {
	Pool  *pgxpool.Pool
	Log   *slog.Logger
	Sales domain.SalesReader
}

// Module is the assembled ledger module.
type Module struct {
	handler *ledgerhttp.Handler
}

// New wires the ledger module.
func New(d Deps) *Module {
	expenses := infra.NewExpenseRepository(d.Pool)
	handler := ledgerhttp.NewHandler(ledgerhttp.HandlerDeps{
		Report: app.NewGetReport(d.Sales, expenses),
		Create: app.NewCreateExpense(expenses),
		List:   app.NewListExpenses(expenses),
		Delete: app.NewDeleteExpense(expenses),
		Log:    d.Log,
	})
	return &Module{handler: handler}
}

// RegisterRoutes mounts the ledger routes onto a JWT-protected router.
func (m *Module) RegisterRoutes(r chi.Router) {
	m.handler.RegisterRoutes(r)
}
