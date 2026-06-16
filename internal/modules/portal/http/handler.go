// Package http is the portal transport layer: staff chat endpoints (/api/app),
// the client portal (/api/portal) and the client-auth middleware.
package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/app"
	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
	"github.com/astralis-s/hakaton-ansar/internal/platform/apperror"
	"github.com/astralis-s/hakaton-ansar/internal/platform/authctx"
	"github.com/astralis-s/hakaton-ansar/internal/platform/pdf"
	"github.com/astralis-s/hakaton-ansar/internal/platform/web"
	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// Handler holds the portal use-cases (staff and client sides).
type Handler struct {
	provision *app.ProvisionAccess
	getAccess *app.GetAccess
	login     *app.LoginClient
	send      *app.SendMessage
	listConv  *app.ListConversations
	thread    *app.GetThread
	profile   *app.GetClientProfile
	contracts *app.GetClientContracts
	contract  *app.GetClientContract
	browse    *app.BrowseProducts
	submitReq *app.SubmitRequest
	myReqs    *app.ListMyRequests
	doc       *app.GetContractDoc
	tgLink    domain.TelegramLinkProvider // optional; set after construction
	log       *slog.Logger
}

// SetTelegramLink wires the manager-deep-link provider (set late, only when the
// Telegram bot module is enabled).
func (h *Handler) SetTelegramLink(p domain.TelegramLinkProvider) { h.tgLink = p }

type HandlerDeps struct {
	Provision *app.ProvisionAccess
	GetAccess *app.GetAccess
	Login     *app.LoginClient
	Send      *app.SendMessage
	ListConv  *app.ListConversations
	Thread    *app.GetThread
	Profile   *app.GetClientProfile
	Contracts *app.GetClientContracts
	Contract  *app.GetClientContract
	Browse    *app.BrowseProducts
	SubmitReq *app.SubmitRequest
	MyReqs    *app.ListMyRequests
	Doc       *app.GetContractDoc
	Log       *slog.Logger
}

func NewHandler(d HandlerDeps) *Handler {
	return &Handler{
		provision: d.Provision, getAccess: d.GetAccess, login: d.Login, send: d.Send,
		listConv: d.ListConv, thread: d.Thread, profile: d.Profile, contracts: d.Contracts,
		contract: d.Contract, browse: d.Browse, submitReq: d.SubmitReq, myReqs: d.MyReqs, doc: d.Doc, log: d.Log,
	}
}

// RegisterStaffRoutes mounts the staff-facing chat routes onto a JWT-protected
// /api/app router.
func (h *Handler) RegisterStaffRoutes(r chi.Router) {
	r.Route("/chats", func(cr chi.Router) {
		cr.Get("/", h.ListChats)
		cr.Get("/telegram-link", h.TelegramLink) // manager's personal bot deep link
		cr.Get("/{clientID}/messages", h.StaffThread)
		cr.Post("/{clientID}/messages", h.StaffSend)
	})
}

// TelegramLink returns the logged-in manager's personal Telegram bot deep link
// for the staff chat page. Available=false when the bot is not configured.
func (h *Handler) TelegramLink(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	if h.tgLink == nil {
		web.JSON(w, http.StatusOK, telegramLinkResponse{Available: false})
		return
	}
	url, qr, available, err := h.tgLink.ManagerLink(r.Context(), p.OrgID, p.UserID)
	if err != nil {
		apperror.Write(w, r, h.log, apperror.Internal("telegram link failed", err))
		return
	}
	web.JSON(w, http.StatusOK, telegramLinkResponse{Available: available, URL: url, QR: qr})
}

// RegisterPublicPortalRoutes mounts the unauthenticated client login route.
func (h *Handler) RegisterPublicPortalRoutes(r chi.Router) {
	r.Post("/auth/login", h.Login)
}

// RegisterProtectedPortalRoutes mounts the client-JWT-protected portal routes.
func (h *Handler) RegisterProtectedPortalRoutes(r chi.Router) {
	r.Get("/me", h.Me)
	r.Get("/contracts", h.MyContracts)
	r.Get("/contracts/{id}", h.MyContract)
	r.Get("/contracts/{id}/pdf", h.MyContractPDF)
	r.Get("/products", h.Products)   // catalog the client may request from
	r.Get("/requests", h.MyRequests) // the client's own requests
	r.Post("/requests", h.SubmitRequest)
	r.Get("/messages", h.MyMessages)
	r.Post("/messages", h.ClientSend)
}

func (h *Handler) Products(w http.ResponseWriter, r *http.Request) {
	p, _ := clientFrom(r.Context())
	products, err := h.browse.Execute(r.Context(), p.OrgID)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	resp := make([]productCardResponse, 0, len(products))
	for _, pr := range products {
		resp = append(resp, toProductCardResponse(pr))
	}
	web.JSON(w, http.StatusOK, resp)
}

func (h *Handler) MyRequests(w http.ResponseWriter, r *http.Request) {
	p, _ := clientFrom(r.Context())
	requests, err := h.myReqs.Execute(r.Context(), p.OrgID, p.ClientID)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	resp := make([]requestViewResponse, 0, len(requests))
	for _, req := range requests {
		resp = append(resp, toRequestViewResponse(req))
	}
	web.JSON(w, http.StatusOK, resp)
}

func (h *Handler) SubmitRequest(w http.ResponseWriter, r *http.Request) {
	p, _ := clientFrom(r.Context())
	var req submitRequestRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	down := money.Zero(money.DefaultCurrency)
	if req.DesiredDownPayment != "" {
		parsed, err := money.FromString(req.DesiredDownPayment, money.DefaultCurrency)
		if err != nil {
			apperror.Write(w, r, h.log, apperror.Invalid("invalid_down_payment", "desired_down_payment must be a decimal string"))
			return
		}
		down = parsed
	}
	view, err := h.submitReq.Execute(r.Context(), app.SubmitRequestInput{
		OrgID:               p.OrgID,
		ClientID:            p.ClientID,
		ProductID:           req.ProductID,
		DesiredInstallments: req.DesiredInstallments,
		DesiredDownPayment:  down,
		Note:                req.Note,
	})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusCreated, toRequestViewResponse(view))
}

// ---- staff side ------------------------------------------------------------

func (h *Handler) ListChats(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	views, names, err := h.listConv.Execute(r.Context(), p.OrgID)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	resp := make([]conversationResponse, 0, len(views))
	for _, v := range views {
		resp = append(resp, toConversationResponse(v, names[v.ClientID]))
	}
	web.JSON(w, http.StatusOK, resp)
}

func (h *Handler) StaffThread(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	messages, err := h.thread.Execute(r.Context(), p.OrgID, chi.URLParam(r, "clientID"))
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, messagesResponse(messages))
}

func (h *Handler) StaffSend(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	var req sendMessageRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	msg, err := h.send.Execute(r.Context(), app.SendMessageInput{
		OrgID:      p.OrgID,
		ClientID:   chi.URLParam(r, "clientID"),
		SenderKind: domain.SenderStaff,
		SenderID:   p.UserID,
		Body:       req.Body,
	})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusCreated, toMessageResponse(msg))
}

func (h *Handler) GetAccess(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	account, err := h.getAccess.Execute(r.Context(), p.OrgID, chi.URLParam(r, "clientID"))
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			web.JSON(w, http.StatusOK, accessResponse{HasAccess: false})
			return
		}
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, accessResponse{HasAccess: true, Email: account.Email()})
}

func (h *Handler) ProvisionAccess(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	var req provisionAccessRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	account, err := h.provision.Execute(r.Context(), app.ProvisionAccessInput{
		OrgID:    p.OrgID,
		ClientID: chi.URLParam(r, "clientID"),
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, accessResponse{HasAccess: true, Email: account.Email()})
}

// ---- client side -----------------------------------------------------------

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	out, err := h.login.Execute(r.Context(), req.Email, req.Password)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, loginResponse{Token: out.Token, ExpiresAt: out.ExpiresAt, ClientID: out.ClientID})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	p, _ := clientFrom(r.Context())
	info, err := h.profile.Execute(r.Context(), p.OrgID, p.ClientID)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, clientProfileResponse{ID: info.ID, FullName: info.FullName, Phone: info.Phone})
}

func (h *Handler) MyContracts(w http.ResponseWriter, r *http.Request) {
	p, _ := clientFrom(r.Context())
	contracts, err := h.contracts.Execute(r.Context(), p.OrgID, p.ClientID)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	resp := make([]contractViewResponse, 0, len(contracts))
	for _, c := range contracts {
		resp = append(resp, toContractViewResponse(c))
	}
	web.JSON(w, http.StatusOK, resp)
}

func (h *Handler) MyContract(w http.ResponseWriter, r *http.Request) {
	p, _ := clientFrom(r.Context())
	detail, err := h.contract.Execute(r.Context(), p.OrgID, p.ClientID, chi.URLParam(r, "id"))
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, toContractDetailResponse(detail))
}

func (h *Handler) MyContractPDF(w http.ResponseWriter, r *http.Request) {
	p, _ := clientFrom(r.Context())
	doc, err := h.doc.Execute(r.Context(), p.OrgID, chi.URLParam(r, "id"), p.ClientID)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	data, err := pdf.RenderContract(doc)
	if err != nil {
		apperror.Write(w, r, h.log, apperror.Internal("render contract pdf", err))
		return
	}
	web.WritePDF(w, "dogovor-"+doc.Number+".pdf", data)
}

func (h *Handler) MyMessages(w http.ResponseWriter, r *http.Request) {
	p, _ := clientFrom(r.Context())
	messages, err := h.thread.Execute(r.Context(), p.OrgID, p.ClientID)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, messagesResponse(messages))
}

func (h *Handler) ClientSend(w http.ResponseWriter, r *http.Request) {
	p, _ := clientFrom(r.Context())
	var req sendMessageRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	msg, err := h.send.Execute(r.Context(), app.SendMessageInput{
		OrgID:      p.OrgID,
		ClientID:   p.ClientID,
		SenderKind: domain.SenderClient,
		SenderID:   p.ClientID,
		Body:       req.Body,
	})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusCreated, toMessageResponse(msg))
}

func messagesResponse(messages []domain.Message) []messageResponse {
	resp := make([]messageResponse, 0, len(messages))
	for _, m := range messages {
		resp = append(resp, toMessageResponse(m))
	}
	return resp
}

func mapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrInvalidCredentials):
		return apperror.Unauthorized("invalid_credentials", "invalid email or password")
	case errors.Is(err, domain.ErrAccountNotFound):
		return apperror.NotFound("account_not_found", "portal account not found")
	case errors.Is(err, domain.ErrClientNotFound):
		return apperror.NotFound("client_not_found", "client not found")
	case errors.Is(err, domain.ErrContractNotFound):
		return apperror.NotFound("contract_not_found", "contract not found")
	case errors.Is(err, domain.ErrProductNotFound):
		return apperror.NotFound("product_not_found", "product not found")
	case errors.Is(err, domain.ErrProductHaram):
		return apperror.Conflict("product_haram", "этот товар нельзя оформить в рассрочку")
	case errors.Is(err, domain.ErrInvalidRequest):
		return apperror.Invalid("invalid_request", "неверные параметры заявки")
	case errors.Is(err, domain.ErrEmailTaken):
		return apperror.Conflict("email_taken", "этот email уже используется")
	case errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrPasswordTooShort),
		errors.Is(err, domain.ErrMessageBodyRequired),
		errors.Is(err, domain.ErrMessageTooLong),
		errors.Is(err, domain.ErrInvalidSenderKind):
		return apperror.Invalid("invalid_input", err.Error())
	default:
		return apperror.Internal("portal operation failed", err)
	}
}
