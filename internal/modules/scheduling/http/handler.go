// Package http is the scheduling transport layer (handlers, DTO, routes).
package http

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/astralis-s/hakaton-ansar/internal/modules/scheduling/app"
	"github.com/astralis-s/hakaton-ansar/internal/modules/scheduling/domain"
	"github.com/astralis-s/hakaton-ansar/internal/platform/apperror"
	"github.com/astralis-s/hakaton-ansar/internal/platform/authctx"
	"github.com/astralis-s/hakaton-ansar/internal/platform/web"
)

// Handler holds the scheduling use-cases.
type Handler struct {
	schedule *app.ScheduleReminder
	preview  *app.PreviewSlot
	list     *app.ListReminders
	get      *app.GetReminder
	update   *app.UpdateReminder
	complete *app.CompleteReminder
	cancel   *app.CancelReminder
	log      *slog.Logger
}

// HandlerDeps groups the use-cases for NewHandler.
type HandlerDeps struct {
	Schedule *app.ScheduleReminder
	Preview  *app.PreviewSlot
	List     *app.ListReminders
	Get      *app.GetReminder
	Update   *app.UpdateReminder
	Complete *app.CompleteReminder
	Cancel   *app.CancelReminder
	Log      *slog.Logger
}

func NewHandler(d HandlerDeps) *Handler {
	return &Handler{
		schedule: d.Schedule, preview: d.Preview, list: d.List, get: d.Get,
		update: d.Update, complete: d.Complete, cancel: d.Cancel, log: d.Log,
	}
}

// RegisterRoutes mounts the scheduling routes (caller provides JWT-protected r).
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/schedule", func(sr chi.Router) {
		sr.Get("/reminders", h.List)
		sr.Post("/reminders", h.Create)
		sr.Get("/reminders/{id}", h.Get)
		sr.Put("/reminders/{id}", h.Update)
		sr.Post("/reminders/{id}/complete", h.Complete)
		sr.Post("/reminders/{id}/cancel", h.Cancel)
		sr.Post("/preview", h.Preview)
	})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	var req createReminderRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	desiredAt, err := time.Parse(time.RFC3339, req.DesiredAt)
	if err != nil {
		apperror.Write(w, r, h.log, apperror.Invalid("invalid_desired_at", "desired_at must be an RFC3339 timestamp"))
		return
	}
	reminder, err := h.schedule.Execute(r.Context(), app.ScheduleReminderInput{
		OrgID: p.OrgID, Type: req.Type, ClientID: req.ClientID, ContractID: req.ContractID,
		Note: req.Note, DesiredAt: desiredAt, DurationMinutes: req.DurationMinutes,
	})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusCreated, toReminderResponse(reminder))
}

func (h *Handler) Preview(w http.ResponseWriter, r *http.Request) {
	var req previewSlotRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	desiredAt, err := time.Parse(time.RFC3339, req.DesiredAt)
	if err != nil {
		apperror.Write(w, r, h.log, apperror.Invalid("invalid_desired_at", "desired_at must be an RFC3339 timestamp"))
		return
	}
	slot, err := h.preview.Execute(r.Context(), app.PreviewSlotInput{DesiredAt: desiredAt, DurationMinutes: req.DurationMinutes})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, toSlotResponse(slot))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	reminders, err := h.list.Execute(r.Context(), p.OrgID)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	resp := make([]reminderResponse, 0, len(reminders))
	for _, rem := range reminders {
		resp = append(resp, toReminderResponse(rem))
	}
	web.JSON(w, http.StatusOK, resp)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	reminder, err := h.get.Execute(r.Context(), p.OrgID, chi.URLParam(r, "id"))
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, toReminderResponse(reminder))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	var req updateReminderRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	desiredAt, err := time.Parse(time.RFC3339, req.DesiredAt)
	if err != nil {
		apperror.Write(w, r, h.log, apperror.Invalid("invalid_desired_at", "desired_at must be an RFC3339 timestamp"))
		return
	}
	reminder, err := h.update.Execute(r.Context(), app.UpdateReminderInput{
		OrgID: p.OrgID, ReminderID: chi.URLParam(r, "id"), Type: req.Type,
		ClientID: req.ClientID, ContractID: req.ContractID, Note: req.Note,
		DesiredAt: desiredAt, DurationMinutes: req.DurationMinutes,
	})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, toReminderResponse(reminder))
}

func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	reminder, err := h.complete.Execute(r.Context(), p.OrgID, chi.URLParam(r, "id"))
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, toReminderResponse(reminder))
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	reminder, err := h.cancel.Execute(r.Context(), p.OrgID, chi.URLParam(r, "id"))
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, toReminderResponse(reminder))
}

func mapError(err error) error {
	var ae *apperror.Error
	if errors.As(err, &ae) {
		return err
	}
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrReminderNotFound):
		return apperror.NotFound("reminder_not_found", "reminder not found")
	case errors.Is(err, domain.ErrInvalidReminderType),
		errors.Is(err, domain.ErrReminderAlreadyDone),
		errors.Is(err, domain.ErrReminderCancelled):
		return apperror.Invalid("invalid_input", err.Error())
	default:
		return apperror.Internal("scheduling operation failed", err)
	}
}
