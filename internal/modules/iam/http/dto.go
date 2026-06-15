package http

import (
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/modules/iam/app"
	"github.com/astralis-s/hakaton-ansar/internal/modules/iam/domain"
)

// --- Requests ---------------------------------------------------------------

type setupRequest struct {
	OrgName       string `json:"org_name" validate:"required"`
	Currency      string `json:"currency"`
	OwnerName     string `json:"owner_name" validate:"required"`
	OwnerEmail    string `json:"owner_email" validate:"required,email"`
	OwnerPassword string `json:"owner_password" validate:"required,min=8"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type createUserRequest struct {
	FullName string `json:"full_name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role" validate:"required,oneof=owner manager"`
}

type createApiKeyRequest struct {
	Name string `json:"name" validate:"required"`
}

// --- Responses --------------------------------------------------------------

type userResponse struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type organizationResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
}

type loginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      userResponse `json:"user"`
}

type setupResponse struct {
	Organization organizationResponse `json:"organization"`
	Owner        userResponse         `json:"owner"`
}

type apiKeyResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Prefix    string     `json:"prefix"`
	Active    bool       `json:"active"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

// createApiKeyResponse includes the plaintext key, returned exactly once.
type createApiKeyResponse struct {
	apiKeyResponse
	Key string `json:"key"`
}

// --- Mappers ----------------------------------------------------------------

func toUserResponse(u domain.User) userResponse {
	return userResponse{
		ID:        u.ID(),
		OrgID:     u.OrgID(),
		FullName:  u.FullName(),
		Email:     u.Email(),
		Role:      u.Role().String(),
		CreatedAt: u.CreatedAt(),
	}
}

func toOrganizationResponse(o domain.Organization) organizationResponse {
	return organizationResponse{
		ID:        o.ID(),
		Name:      o.Name(),
		Currency:  o.Currency(),
		CreatedAt: o.CreatedAt(),
	}
}

func toAPIKeyResponse(k domain.ApiKey) apiKeyResponse {
	return apiKeyResponse{
		ID:        k.ID(),
		Name:      k.Name(),
		Prefix:    k.Prefix(),
		Active:    k.IsActive(),
		CreatedAt: k.CreatedAt(),
		RevokedAt: k.RevokedAt(),
	}
}

func toCreateAPIKeyResponse(out app.CreateApiKeyOutput) createApiKeyResponse {
	return createApiKeyResponse{
		apiKeyResponse: toAPIKeyResponse(out.Key),
		Key:            out.PlainKey,
	}
}
