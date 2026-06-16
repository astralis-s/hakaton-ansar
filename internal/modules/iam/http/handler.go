package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/astralis-s/hakaton-ansar/internal/modules/iam/app"
	"github.com/astralis-s/hakaton-ansar/internal/modules/iam/domain"
	"github.com/astralis-s/hakaton-ansar/internal/platform/apperror"
	"github.com/astralis-s/hakaton-ansar/internal/platform/authctx"
	"github.com/astralis-s/hakaton-ansar/internal/platform/web"
)

// Handler holds the iam use-cases and maps DTO ↔ domain.
type Handler struct {
	setup      *app.SetupOrganization
	register   *app.RegisterOrganization
	login      *app.Login
	createUser *app.CreateUser
	listUsers  *app.ListUsers
	getUser    *app.GetUser
	createKey  *app.CreateApiKey
	listKeys   *app.ListApiKeys
	revokeKey  *app.RevokeApiKey
	log        *slog.Logger
}

func (h *Handler) Setup(w http.ResponseWriter, r *http.Request) {
	var req setupRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	out, err := h.setup.Execute(r.Context(), app.SetupInput{
		OrgName:       req.OrgName,
		Currency:      req.Currency,
		OwnerName:     req.OwnerName,
		OwnerEmail:    req.OwnerEmail,
		OwnerPassword: req.OwnerPassword,
	})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusCreated, setupResponse{
		Organization: toOrganizationResponse(out.Org),
		Owner:        toUserResponse(out.Owner),
	})
}

// Register creates a new organization with its owner (multi-tenant sign-up). It
// works at any time (Amana is multi-organization), unlike first-run Setup.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req setupRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	out, err := h.register.Execute(r.Context(), app.RegisterInput{
		OrgName:       req.OrgName,
		Currency:      req.Currency,
		OwnerName:     req.OwnerName,
		OwnerEmail:    req.OwnerEmail,
		OwnerPassword: req.OwnerPassword,
	})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusCreated, setupResponse{
		Organization: toOrganizationResponse(out.Org),
		Owner:        toUserResponse(out.Owner),
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	out, err := h.login.Execute(r.Context(), app.LoginInput{Email: req.Email, Password: req.Password})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, loginResponse{
		Token:     out.Token,
		ExpiresAt: out.ExpiresAt,
		User:      toUserResponse(out.User),
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	p, ok := authctx.From(r.Context())
	if !ok {
		apperror.Write(w, r, h.log, apperror.Unauthorized("unauthenticated", "authentication required"))
		return
	}
	user, err := h.getUser.Execute(r.Context(), p.UserID)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, toUserResponse(user))
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	var req createUserRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	user, err := h.createUser.Execute(r.Context(), app.CreateUserInput{
		OrgID:    p.OrgID,
		FullName: req.FullName,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusCreated, toUserResponse(user))
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	users, err := h.listUsers.Execute(r.Context(), p.OrgID)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	resp := make([]userResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, toUserResponse(u))
	}
	web.JSON(w, http.StatusOK, resp)
}

func (h *Handler) CreateApiKey(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	var req createApiKeyRequest
	if err := web.DecodeAndValidate(w, r, &req); err != nil {
		apperror.Write(w, r, h.log, err)
		return
	}
	out, err := h.createKey.Execute(r.Context(), app.CreateApiKeyInput{OrgID: p.OrgID, Name: req.Name})
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusCreated, toCreateAPIKeyResponse(out))
}

func (h *Handler) ListApiKeys(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	keys, err := h.listKeys.Execute(r.Context(), p.OrgID)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	resp := make([]apiKeyResponse, 0, len(keys))
	for _, k := range keys {
		resp = append(resp, toAPIKeyResponse(k))
	}
	web.JSON(w, http.StatusOK, resp)
}

func (h *Handler) RevokeApiKey(w http.ResponseWriter, r *http.Request) {
	p, _ := authctx.From(r.Context())
	id := chi.URLParam(r, "id")
	key, err := h.revokeKey.Execute(r.Context(), id, p.OrgID)
	if err != nil {
		apperror.Write(w, r, h.log, mapError(err))
		return
	}
	web.JSON(w, http.StatusOK, toAPIKeyResponse(key))
}

// mapError classifies iam domain errors into HTTP-aware apperrors.
func mapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrInvalidCredentials):
		return apperror.Unauthorized("invalid_credentials", "invalid email or password")
	case errors.Is(err, domain.ErrUserNotFound):
		return apperror.NotFound("user_not_found", "user not found")
	case errors.Is(err, domain.ErrOrgNotFound):
		return apperror.NotFound("organization_not_found", "organization not found")
	case errors.Is(err, domain.ErrApiKeyNotFound):
		return apperror.NotFound("api_key_not_found", "API key not found")
	case errors.Is(err, domain.ErrEmailTaken):
		return apperror.Conflict("email_taken", "email already in use")
	case errors.Is(err, domain.ErrAlreadyInitialized):
		return apperror.Conflict("already_initialized", "organization already initialized")
	case errors.Is(err, domain.ErrInvalidRole),
		errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrOrgNameRequired),
		errors.Is(err, domain.ErrFullNameRequired),
		errors.Is(err, domain.ErrPasswordHashEmpty),
		errors.Is(err, domain.ErrApiKeyNameRequired):
		return apperror.Invalid("invalid_input", err.Error())
	default:
		return apperror.Internal("iam operation failed", err)
	}
}
