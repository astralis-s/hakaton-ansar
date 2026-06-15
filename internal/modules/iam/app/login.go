package app

import (
	"context"
	"errors"
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/modules/iam/domain"
)

// Login authenticates a user by email + password and issues a JWT for the
// internal API.
type Login struct {
	users  domain.UserRepository
	hasher domain.PasswordHasher
	tokens domain.TokenService
}

func NewLogin(users domain.UserRepository, hasher domain.PasswordHasher, tokens domain.TokenService) *Login {
	return &Login{users: users, hasher: hasher, tokens: tokens}
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	Token     string
	ExpiresAt time.Time
	User      domain.User
}

func (uc *Login) Execute(ctx context.Context, in LoginInput) (LoginOutput, error) {
	user, err := uc.users.GetByEmail(ctx, in.Email)
	if err != nil {
		// Do not reveal whether the email exists.
		if errors.Is(err, domain.ErrUserNotFound) {
			return LoginOutput{}, domain.ErrInvalidCredentials
		}
		return LoginOutput{}, err
	}

	if err := uc.hasher.Compare(user.PasswordHash(), in.Password); err != nil {
		return LoginOutput{}, domain.ErrInvalidCredentials
	}

	token, expiresAt, err := uc.tokens.Issue(user)
	if err != nil {
		return LoginOutput{}, err
	}
	return LoginOutput{Token: token, ExpiresAt: expiresAt, User: user}, nil
}
