// Package financing is the composition root of the financing bounded context
// (the murabaha core).
package financing

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/app"
	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
	financinghttp "github.com/astralis-s/hakaton-ansar/internal/modules/financing/http"
	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/infra"
)

// Deps are the external dependencies of the financing module. Products/Clients
// are cross-context reader ports (wired from catalog/crm in main); OwnerOnly is
// the middleware gating owner-only actions (cancel, charity accrual).
type Deps struct {
	Pool                  *pgxpool.Pool
	Tx                    domain.TxManager
	Log                   *slog.Logger
	ComparisonRatePercent decimal.Decimal
	Products              domain.ProductReader
	Clients               domain.ClientReader
	OwnerOnly             func(http.Handler) http.Handler
}

// Module is the assembled financing module.
type Module struct {
	handler        *financinghttp.Handler
	createContract *app.CreateContract
	getContract    *app.GetContract
}

// New wires the financing module.
func New(d Deps) *Module {
	contracts := infra.NewContractRepository(d.Pool)
	charity := infra.NewCharityRepository(d.Pool)

	createContract := app.NewCreateContract(contracts, d.Products, d.Clients, d.Tx)
	getContract := app.NewGetContract(contracts)

	handler := financinghttp.NewHandler(financinghttp.HandlerDeps{
		Preview:     app.NewPreviewContract(d.ComparisonRatePercent),
		Create:      createContract,
		Get:         getContract,
		List:        app.NewListContracts(contracts),
		Pay:         app.NewRegisterPayment(contracts, d.Tx),
		Settle:      app.NewSettleEarly(contracts, d.Tx),
		Cancel:      app.NewCancelContract(contracts, d.Tx),
		Accrue:      app.NewAccrueLateCharity(contracts, charity),
		ListCharity: app.NewListCharity(charity),
		Log:         d.Log,
		OwnerOnly:   d.OwnerOnly,
	})
	return &Module{handler: handler, createContract: createContract, getContract: getContract}
}

// RegisterRoutes mounts the financing routes onto a JWT-protected router.
func (m *Module) RegisterRoutes(r chi.Router) {
	m.handler.RegisterRoutes(r)
}

// CreateContractUseCase exposes the create use-case for the public API (thin
// layer over the same application service — no logic duplication).
func (m *Module) CreateContractUseCase() *app.CreateContract { return m.createContract }

// GetContractUseCase exposes the get use-case for the public API.
func (m *Module) GetContractUseCase() *app.GetContract { return m.getContract }
