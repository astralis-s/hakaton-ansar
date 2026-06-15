// Package http is the crm transport layer (handlers, DTO, routes).
package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/astralis-s/hakaton-ansar/internal/modules/crm/app"
	"github.com/astralis-s/hakaton-ansar/internal/modules/crm/domain"
	"github.com/astralis-s/hakaton-ansar/internal/platform/apperror"
	"github.com/astralis-s/hakaton-ansar/internal/platform/authctx"
	"github.com/astralis-s/hakaton-ansar/internal/platform/web"
)

// Handler holds the crm use-cases.
type Handler struct {
	create *app.CreateClient
	get    *app.GetClient
	list   *app.ListClients
	update *app.UpdateClient
	log    *slog.Logger
}

// HandlerDeps groups the use-cases for NewHandler.
type HandlerDeps struct {
	Create *app.CreateClient
	Get    *app.GetClient
	List   *app.ListClients
	Update *app.UpdateClient
	Log    *slog.Logger
}

func NewHandler(d HandlerDeps) *Handler {
	return &Handler{create: d.Create, get: d.Get, list: d.List, update: d.Update, log: d.Log}
}

// RegisterRoutes mounts the crm routes (caller provides JWT-protected r).
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/clients", func(cr chi.Router) {
		cr.Get("/", h.List)
		cr.Post("/", h.Create)
		cr.Get("/{id}", h.Get)
		cr.Put("/{id}", h.Update)
	})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	var req createClientRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	client, err := h.create.Execute(r.Context(), app.CreateClientInput{
		OrgID:    p.OrgID,
		FullName: req.FullName,
		Phone:    req.Phone,
		Document: req.Document,
	})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusCreated, toClientResponse(client))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	client, err := h.get.Execute(r.Context(), p.OrgID, chi.URLParam(r, "id"))
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, toClientResponse(client))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	clients, err := h.list.Execute(r.Context(), p.OrgID)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	resp := make([]clientResponse, 0, len(clients))
	for _, client := range clients {
		resp = append(resp, toClientResponse(client))
	}
	web.JSON(w, http.StatusOK, resp)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	var req updateClientRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	client, err := h.update.Execute(r.Context(), app.UpdateClientInput{
		OrgID:    p.OrgID,
		ID:       chi.URLParam(r, "id"),
		FullName: req.FullName,
		Phone:    req.Phone,
		Document: req.Document,
	})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, toClientResponse(client))
}

func mapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrClientNotFound):
		return apperror.NotFound("client_not_found", "client not found")
	case errors.Is(err, domain.ErrClientNameRequired):
		return apperror.Invalid("invalid_input", err.Error())
	default:
		return apperror.Internal("crm operation failed", err)
	}
}
