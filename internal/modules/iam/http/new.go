package http

import (
	"log/slog"

	"github.com/astralis-s/hakaton-ansar/internal/modules/iam/app"
)

// HandlerDeps groups the use-cases a Handler needs.
type HandlerDeps struct {
	Setup      *app.SetupOrganization
	Login      *app.Login
	CreateUser *app.CreateUser
	ListUsers  *app.ListUsers
	GetUser    *app.GetUser
	CreateKey  *app.CreateApiKey
	ListKeys   *app.ListApiKeys
	RevokeKey  *app.RevokeApiKey
	Log        *slog.Logger
}

// NewHandler assembles the iam HTTP handler.
func NewHandler(d HandlerDeps) *Handler {
	return &Handler{
		setup:      d.Setup,
		login:      d.Login,
		createUser: d.CreateUser,
		listUsers:  d.ListUsers,
		getUser:    d.GetUser,
		createKey:  d.CreateKey,
		listKeys:   d.ListKeys,
		revokeKey:  d.RevokeKey,
		log:        d.Log,
	}
}
