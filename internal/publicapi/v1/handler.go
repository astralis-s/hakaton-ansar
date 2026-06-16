// Package v1 is the public API (/api/v1) — a thin transport layer over the same
// financing application services (it does not duplicate business logic). It is
// authenticated by X-API-Key (middleware applied where it is mounted).
package v1

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	financingapp "github.com/astralis-s/hakaton-ansar/internal/modules/financing/app"
	financingdomain "github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
	"github.com/astralis-s/hakaton-ansar/internal/platform/apperror"
	"github.com/astralis-s/hakaton-ansar/internal/platform/authctx"
	"github.com/astralis-s/hakaton-ansar/internal/platform/web"
	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// Handler serves the public API endpoints over the financing use-cases.
type Handler struct {
	create *financingapp.CreateContract
	get    *financingapp.GetContract
	log    *slog.Logger
}

// NewHandler builds the public API handler.
func NewHandler(create *financingapp.CreateContract, get *financingapp.GetContract, log *slog.Logger) *Handler {
	return &Handler{create: create, get: get, log: log}
}

// RegisterRoutes mounts the public API routes (X-API-Key middleware is applied by
// the caller).
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/contracts", func(cr chi.Router) {
		cr.Post("/", h.CreateContract)
		cr.Get("/{id}/payments", h.GetPayments)
	})
}

// CreateContract godoc
//
//	@Summary		Create an installment contract (murabaha)
//	@Description	Creates an installment contract referencing an existing client and product. The sale price is fixed at creation (cost + markup); the debt never grows with time. Money fields are decimal strings (e.g. "120000.00").
//	@Tags			contracts
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CreateContractRequest	true	"Contract terms"
//	@Success		201		{object}	ContractResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Security		ApiKeyAuth
//	@Router			/contracts [post]
func (h *Handler) CreateContract(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())

	var req CreateContractRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}

	cost, err := money.FromString(req.CostPrice, money.DefaultCurrency)
	if err != nil {
		apperror.Write(w, r, h.log, apperror.Invalid("invalid_cost_price", "cost_price must be a decimal string"))
		return
	}
	markupMoney, err := money.FromString(req.Markup, money.DefaultCurrency)
	if err != nil {
		apperror.Write(w, r, h.log, apperror.Invalid("invalid_markup", "markup must be a decimal string"))
		return
	}
	markup, err := financingdomain.NewMarkup(markupMoney)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	downStr := req.DownPayment
	if downStr == "" {
		downStr = "0"
	}
	down, err := money.FromString(downStr, money.DefaultCurrency)
	if err != nil {
		apperror.Write(w, r, h.log, apperror.Invalid("invalid_down_payment", "down_payment must be a decimal string"))
		return
	}
	cadence, err := financingdomain.ParseCadence(req.Cadence)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	start, err := time.Parse(dateLayout, req.StartDate)
	if err != nil {
		apperror.Write(w, r, h.log, apperror.Invalid("invalid_start_date", "start_date must be YYYY-MM-DD"))
		return
	}

	contract, err := h.create.Execute(r.Context(), financingapp.CreateContractInput{
		OrgID:        p.OrgID,
		ClientID:     req.ClientID,
		ProductID:    req.ProductID,
		CostPrice:    cost,
		Markup:       markup,
		DownPayment:  down,
		Installments: req.Installments,
		Cadence:      cadence,
		StartDate:    start,
	})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusCreated, toContractResponse(contract))
}

// GetPayments godoc
//
//	@Summary		Get a contract's payment status and schedule
//	@Description	Returns the outstanding balance, progress and the full installment schedule (with derived statuses) plus registered payments.
//	@Tags			contracts
//	@Produce		json
//	@Param			id	path		string	true	"Contract ID"
//	@Success		200	{object}	PaymentsResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Security		ApiKeyAuth
//	@Router			/contracts/{id}/payments [get]
func (h *Handler) GetPayments(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	contract, err := h.get.Execute(r.Context(), p.OrgID, chi.URLParam(r, "id"))
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, toPaymentsResponse(contract, time.Now()))
}

// mapError classifies financing errors for the public API. Already-classified
// apperrors pass through unchanged.
func mapError(err error) error {
	var ae *apperror.Error
	if errors.As(err, &ae) {
		return err
	}
	switch {
	case err == nil:
		return nil
	case errors.Is(err, financingdomain.ErrContractNotFound):
		return apperror.NotFound("contract_not_found", "contract not found")
	case errors.Is(err, financingdomain.ErrProductNotFound):
		return apperror.NotFound("product_not_found", "product not found")
	case errors.Is(err, financingdomain.ErrClientNotFound):
		return apperror.NotFound("client_not_found", "client not found")
	case errors.Is(err, financingdomain.ErrProductHaram):
		return apperror.Conflict("product_haram", "cannot create a contract for a haram product")
	case errors.Is(err, financingdomain.ErrOutOfStock):
		return apperror.Conflict("out_of_stock", "product is out of stock")
	case errors.Is(err, financingdomain.ErrDownPaymentTooLarge),
		errors.Is(err, financingdomain.ErrDownPaymentNegative),
		errors.Is(err, financingdomain.ErrInstallmentsNotPositive),
		errors.Is(err, financingdomain.ErrFinancedLessThanInstallment),
		errors.Is(err, financingdomain.ErrCostPriceNotPositive),
		errors.Is(err, financingdomain.ErrMarkupNegative),
		errors.Is(err, financingdomain.ErrInvalidCadence):
		return apperror.Invalid("invalid_input", err.Error())
	default:
		return apperror.Internal("public api operation failed", err)
	}
}
