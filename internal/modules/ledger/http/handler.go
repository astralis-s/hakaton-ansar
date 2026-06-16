// Package http is the ledger transport layer (handlers, DTO, routes).
package http

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/astralis-s/hakaton-ansar/internal/modules/ledger/app"
	"github.com/astralis-s/hakaton-ansar/internal/modules/ledger/domain"
	"github.com/astralis-s/hakaton-ansar/internal/platform/apperror"
	"github.com/astralis-s/hakaton-ansar/internal/platform/authctx"
	"github.com/astralis-s/hakaton-ansar/internal/platform/pdf"
	"github.com/astralis-s/hakaton-ansar/internal/platform/web"
	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// Handler holds the ledger use-cases.
type Handler struct {
	report    *app.GetReport
	create    *app.CreateExpense
	list      *app.ListExpenses
	del       *app.DeleteExpense
	reportDoc *app.BuildReportDoc
	log       *slog.Logger
}

// HandlerDeps groups the use-cases for NewHandler.
type HandlerDeps struct {
	Report    *app.GetReport
	Create    *app.CreateExpense
	List      *app.ListExpenses
	Delete    *app.DeleteExpense
	ReportDoc *app.BuildReportDoc
	Log       *slog.Logger
}

func NewHandler(d HandlerDeps) *Handler {
	return &Handler{report: d.Report, create: d.Create, list: d.List, del: d.Delete, reportDoc: d.ReportDoc, log: d.Log}
}

// RegisterRoutes mounts the ledger routes (caller provides JWT-protected r).
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/finance", func(fr chi.Router) {
		fr.Get("/report", h.Report)       // P&L summary + per-sale income breakdown
		fr.Get("/report.pdf", h.ReportPDF) // one-page P&L report
		fr.Get("/expenses", h.ListExpenses)
		fr.Post("/expenses", h.CreateExpense)
		fr.Delete("/expenses/{id}", h.DeleteExpense)
	})
}

// ReportPDF streams the one-page finance report as a PDF.
func (h *Handler) ReportPDF(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	doc, err := h.reportDoc.Execute(r.Context(), p.OrgID)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	data, err := pdf.RenderFinanceReport(doc)
	if err != nil {
		apperror.Write(w, r, h.log, apperror.Internal("render finance report pdf", err))
		return
	}
	web.WritePDF(w, "finansy-otchet.pdf", data)
}

func (h *Handler) Report(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	report, sales, err := h.report.Execute(r.Context(), p.OrgID)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, toReportResponse(report, sales))
}

func (h *Handler) ListExpenses(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	expenses, err := h.list.Execute(r.Context(), p.OrgID)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	resp := make([]expenseResponse, 0, len(expenses))
	for _, e := range expenses {
		resp = append(resp, toExpenseResponse(e))
	}
	web.JSON(w, http.StatusOK, resp)
}

func (h *Handler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	var req createExpenseRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	amount, err := money.FromString(req.Amount, money.DefaultCurrency)
	if err != nil {
		apperror.Write(w, r, h.log, apperror.Invalid("invalid_amount", "amount must be a decimal string"))
		return
	}
	var spentAt time.Time
	if req.SpentAt != "" {
		spentAt, err = time.Parse(dateLayout, req.SpentAt)
		if err != nil {
			apperror.Write(w, r, h.log, apperror.Invalid("invalid_spent_at", "spent_at must be YYYY-MM-DD"))
			return
		}
	}
	expense, err := h.create.Execute(r.Context(), app.CreateExpenseInput{
		OrgID:    p.OrgID,
		Category: req.Category,
		Amount:   amount,
		Note:     req.Note,
		SpentAt:  spentAt,
	})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusCreated, toExpenseResponse(expense))
}

func (h *Handler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	if err := h.del.Execute(r.Context(), p.OrgID, chi.URLParam(r, "id")); err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusNoContent, nil)
}

func mapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrExpenseNotFound):
		return apperror.NotFound("expense_not_found", "expense not found")
	case errors.Is(err, domain.ErrCategoryRequired),
		errors.Is(err, domain.ErrAmountNotPositive):
		return apperror.Invalid("invalid_input", err.Error())
	default:
		return apperror.Internal("ledger operation failed", err)
	}
}
