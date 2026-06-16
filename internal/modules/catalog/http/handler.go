// Package http is the catalog transport layer (handlers, DTO, routes).
package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/astralis-s/hakaton-ansar/internal/modules/catalog/app"
	"github.com/astralis-s/hakaton-ansar/internal/modules/catalog/domain"
	"github.com/astralis-s/hakaton-ansar/internal/platform/apperror"
	"github.com/astralis-s/hakaton-ansar/internal/platform/authctx"
	"github.com/astralis-s/hakaton-ansar/internal/platform/web"
	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// Handler holds the catalog use-cases.
type Handler struct {
	create    *app.CreateProduct
	get       *app.GetProduct
	list      *app.ListProducts
	update    *app.UpdateProduct
	adjust    *app.AdjustStock
	movements *app.ListStockMovements
	log       *slog.Logger
}

// HandlerDeps groups the use-cases for NewHandler.
type HandlerDeps struct {
	Create    *app.CreateProduct
	Get       *app.GetProduct
	List      *app.ListProducts
	Update    *app.UpdateProduct
	Adjust    *app.AdjustStock
	Movements *app.ListStockMovements
	Log       *slog.Logger
}

func NewHandler(d HandlerDeps) *Handler {
	return &Handler{create: d.Create, get: d.Get, list: d.List, update: d.Update,
		adjust: d.Adjust, movements: d.Movements, log: d.Log}
}

// RegisterRoutes mounts the catalog routes (caller provides JWT-protected r).
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/catalog", func(cr chi.Router) {
		cr.Get("/", h.List)
		cr.Post("/", h.Create)
		cr.Get("/movements", h.ListMovements) // товарооборот (org-wide)
		cr.Get("/{id}", h.Get)
		cr.Put("/{id}", h.Update)
		cr.Post("/{id}/stock", h.AdjustStock) // receipt / adjustment / write-off
	})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	var req createProductRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	cost, err := money.FromString(req.CostPrice, money.DefaultCurrency)
	if err != nil {
		apperror.Write(w, r, h.log, apperror.Invalid("invalid_cost_price", "cost_price must be a decimal string"))
		return
	}
	product, err := h.create.Execute(r.Context(), app.CreateProductInput{
		OrgID:       p.OrgID,
		Name:        req.Name,
		Category:    req.Category,
		CostPrice:   cost,
		HalalStatus: req.HalalStatus,
		Stock:       req.Stock,
	})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusCreated, toProductResponse(product))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	product, err := h.get.Execute(r.Context(), p.OrgID, chi.URLParam(r, "id"))
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, toProductResponse(product))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	products, err := h.list.Execute(r.Context(), p.OrgID)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	resp := make([]productResponse, 0, len(products))
	for _, product := range products {
		resp = append(resp, toProductResponse(product))
	}
	web.JSON(w, http.StatusOK, resp)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	var req updateProductRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	cost, err := money.FromString(req.CostPrice, money.DefaultCurrency)
	if err != nil {
		apperror.Write(w, r, h.log, apperror.Invalid("invalid_cost_price", "cost_price must be a decimal string"))
		return
	}
	product, err := h.update.Execute(r.Context(), app.UpdateProductInput{
		OrgID:       p.OrgID,
		ID:          chi.URLParam(r, "id"),
		Name:        req.Name,
		Category:    req.Category,
		CostPrice:   cost,
		HalalStatus: req.HalalStatus,
	})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, toProductResponse(product))
}

func (h *Handler) AdjustStock(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	var req adjustStockRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	product, _, err := h.adjust.Execute(r.Context(), app.AdjustStockInput{
		OrgID:     p.OrgID,
		ProductID: chi.URLParam(r, "id"),
		Delta:     req.Delta,
		Reason:    req.Reason,
		Note:      req.Note,
	})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, toProductResponse(product))
}

func (h *Handler) ListMovements(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	movements, err := h.movements.Execute(r.Context(), p.OrgID)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	resp := make([]stockMovementResponse, 0, len(movements))
	for _, m := range movements {
		resp = append(resp, toStockMovementResponse(m))
	}
	web.JSON(w, http.StatusOK, resp)
}

func mapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrProductNotFound):
		return apperror.NotFound("product_not_found", "product not found")
	case errors.Is(err, domain.ErrInsufficientStock):
		return apperror.Conflict("insufficient_stock", "недостаточно товара на складе")
	case errors.Is(err, domain.ErrInvalidHalalStatus),
		errors.Is(err, domain.ErrCostPriceNotPositive),
		errors.Is(err, domain.ErrProductNameRequired),
		errors.Is(err, domain.ErrNegativeStock),
		errors.Is(err, domain.ErrStockDeltaZero),
		errors.Is(err, domain.ErrInvalidStockReason):
		return apperror.Invalid("invalid_input", err.Error())
	default:
		return apperror.Internal("catalog operation failed", err)
	}
}
