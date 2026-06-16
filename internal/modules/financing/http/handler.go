// Package http is the financing transport layer (handlers, DTO, routes).
package http

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"

	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/app"
	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
	"github.com/astralis-s/hakaton-ansar/internal/platform/apperror"
	"github.com/astralis-s/hakaton-ansar/internal/platform/authctx"
	"github.com/astralis-s/hakaton-ansar/internal/platform/web"
	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// Handler holds the financing use-cases.
type Handler struct {
	preview   *app.PreviewContract
	create    *app.CreateContract
	get       *app.GetContract
	list      *app.ListContracts
	pay       *app.RegisterPayment
	settle    *app.SettleEarly
	cancel    *app.CancelContract
	dashboard *app.Dashboard
	log       *slog.Logger
	ownerMW   func(http.Handler) http.Handler
}

// HandlerDeps groups the use-cases for NewHandler.
type HandlerDeps struct {
	Preview   *app.PreviewContract
	Create    *app.CreateContract
	Get       *app.GetContract
	List      *app.ListContracts
	Pay       *app.RegisterPayment
	Settle    *app.SettleEarly
	Cancel    *app.CancelContract
	Dashboard *app.Dashboard
	Log       *slog.Logger
	OwnerOnly func(http.Handler) http.Handler
}

func NewHandler(d HandlerDeps) *Handler {
	return &Handler{
		preview: d.Preview, create: d.Create, get: d.Get, list: d.List,
		pay: d.Pay, settle: d.Settle, cancel: d.Cancel,
		dashboard: d.Dashboard, log: d.Log, ownerMW: d.OwnerOnly,
	}
}

// RegisterRoutes mounts the financing routes (caller provides JWT-protected r).
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/dashboard", h.Dashboard) // owner/manager morning view
	r.Route("/contracts", func(cr chi.Router) {
		cr.Post("/preview", h.Preview) // static path wins over /{id}
		cr.Get("/", h.List)
		cr.Post("/", h.Create)
		cr.Get("/{id}", h.Get)
		cr.Post("/{id}/payments", h.RegisterPayment)
		cr.Post("/{id}/settle", h.SettleEarly)
		cr.With(h.ownerMW).Post("/{id}/cancel", h.Cancel)
	})
}

// Dashboard returns the aggregated owner/manager dashboard.
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	res, err := h.dashboard.Execute(r.Context(), p.OrgID)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, toDashboardResponse(res))
}

func (h *Handler) Preview(w http.ResponseWriter, r *http.Request) {
	var req previewRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	cost, markup, down, cadence, start, err := parseTerms(req.CostPrice, req.MarkupAmount, req.MarkupPercent, req.DownPayment, req.Cadence, req.StartDate)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	result, err := h.preview.Execute(r.Context(), domain.PreviewInput{
		CostPrice: cost, Markup: markup, DownPayment: down,
		Installments: req.Installments, Cadence: cadence, StartDate: start,
	})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, toPreviewResponse(result))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	var req createContractRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	cost, markup, down, cadence, start, err := parseTerms(req.CostPrice, req.MarkupAmount, req.MarkupPercent, req.DownPayment, req.Cadence, req.StartDate)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	contract, err := h.create.Execute(r.Context(), app.CreateContractInput{
		OrgID: p.OrgID, ClientID: req.ClientID, ProductID: req.ProductID,
		CostPrice: cost, Markup: markup, DownPayment: down,
		Installments: req.Installments, Cadence: cadence, StartDate: start,
	})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusCreated, toContractResponse(contract, time.Now()))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	contract, err := h.get.Execute(r.Context(), p.OrgID, chi.URLParam(r, "id"))
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, toContractResponse(contract, time.Now()))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	summaries, err := h.list.Execute(r.Context(), p.OrgID)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	resp := make([]contractSummaryDTO, 0, len(summaries))
	for _, s := range summaries {
		resp = append(resp, toContractSummaryDTO(s))
	}
	web.JSON(w, http.StatusOK, resp)
}

func (h *Handler) RegisterPayment(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	var req registerPaymentRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	amount, err := money.FromString(req.Amount, money.DefaultCurrency)
	if err != nil {
		apperror.Write(w, r, h.log, apperror.Invalid("invalid_amount", "amount must be a decimal string"))
		return
	}
	contract, err := h.pay.Execute(r.Context(), app.RegisterPaymentInput{
		OrgID: p.OrgID, ContractID: chi.URLParam(r, "id"), Amount: amount,
	})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, toContractResponse(contract, time.Now()))
}

func (h *Handler) SettleEarly(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	contract, err := h.settle.Execute(r.Context(), p.OrgID, chi.URLParam(r, "id"))
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, toContractResponse(contract, time.Now()))
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	contract, err := h.cancel.Execute(r.Context(), p.OrgID, chi.URLParam(r, "id"))
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, toContractResponse(contract, time.Now()))
}

// parseTerms converts the raw request strings into domain value objects. Markup
// is supplied as an amount OR a percent of the cost (exactly one).
func parseTerms(costStr, markupAmountStr, markupPercentStr, downStr, cadenceStr, startStr string) (cost money.Money, markup domain.Markup, down money.Money, cadence domain.Cadence, start time.Time, err error) {
	cost, e := money.FromString(costStr, money.DefaultCurrency)
	if e != nil {
		return cost, markup, down, cadence, start, apperror.Invalid("invalid_cost_price", "cost_price must be a decimal string")
	}

	switch {
	case markupPercentStr != "" && markupAmountStr != "":
		return cost, markup, down, cadence, start, apperror.Invalid("markup_ambiguous", "provide either markup_amount or markup_percent, not both")
	case markupPercentStr != "":
		percent, perr := decimal.NewFromString(markupPercentStr)
		if perr != nil {
			return cost, markup, down, cadence, start, apperror.Invalid("invalid_markup_percent", "markup_percent must be a decimal")
		}
		markup, err = domain.NewMarkupFromPercent(cost, percent)
	default:
		amt := markupAmountStr
		if amt == "" {
			amt = "0"
		}
		mm, merr := money.FromString(amt, money.DefaultCurrency)
		if merr != nil {
			return cost, markup, down, cadence, start, apperror.Invalid("invalid_markup_amount", "markup_amount must be a decimal string")
		}
		markup, err = domain.NewMarkup(mm)
	}
	if err != nil {
		return cost, markup, down, cadence, start, err
	}

	downAmt := downStr
	if downAmt == "" {
		downAmt = "0"
	}
	down, e = money.FromString(downAmt, money.DefaultCurrency)
	if e != nil {
		return cost, markup, down, cadence, start, apperror.Invalid("invalid_down_payment", "down_payment must be a decimal string")
	}

	cadence, err = domain.ParseCadence(cadenceStr)
	if err != nil {
		return cost, markup, down, cadence, start, err
	}

	start, e = time.Parse(dateLayout, startStr)
	if e != nil {
		return cost, markup, down, cadence, start, apperror.Invalid("invalid_start_date", "start_date must be YYYY-MM-DD")
	}
	return cost, markup, down, cadence, start, nil
}

// mapError classifies financing errors into HTTP-aware apperrors. Already-classified
// apperrors (from parsing) pass through unchanged.
func mapError(err error) error {
	var ae *apperror.Error
	if errors.As(err, &ae) {
		return err
	}
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrContractNotFound):
		return apperror.NotFound("contract_not_found", "contract not found")
	case errors.Is(err, domain.ErrProductNotFound):
		return apperror.NotFound("product_not_found", "product not found")
	case errors.Is(err, domain.ErrClientNotFound):
		return apperror.NotFound("client_not_found", "client not found")
	case errors.Is(err, domain.ErrProductHaram):
		return apperror.Conflict("product_haram", "cannot create a contract for a haram product")
	case errors.Is(err, domain.ErrOutOfStock):
		return apperror.Conflict("out_of_stock", "товара нет на складе")
	case errors.Is(err, domain.ErrContractNotActive),
		errors.Is(err, domain.ErrInvalidStatusTransition),
		errors.Is(err, domain.ErrAlreadySettled):
		return apperror.Conflict("invalid_state", err.Error())
	case errors.Is(err, domain.ErrPaymentExceedsOutstanding),
		errors.Is(err, domain.ErrPaymentNotPositive),
		errors.Is(err, domain.ErrDownPaymentTooLarge),
		errors.Is(err, domain.ErrDownPaymentNegative),
		errors.Is(err, domain.ErrInstallmentsNotPositive),
		errors.Is(err, domain.ErrFinancedLessThanInstallment),
		errors.Is(err, domain.ErrCostPriceNotPositive),
		errors.Is(err, domain.ErrMarkupNegative),
		errors.Is(err, domain.ErrInvalidCadence):
		return apperror.Invalid("invalid_input", err.Error())
	default:
		return apperror.Internal("financing operation failed", err)
	}
}
