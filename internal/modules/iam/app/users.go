package app

import (
	"context"

	"github.com/astralis-s/hakaton-ansar/internal/modules/iam/domain"
)

// CreateUser lets an owner add a user (owner or manager) to their organization.
type CreateUser struct {
	users  domain.UserRepository
	hasher domain.PasswordHasher
}

func NewCreateUser(users domain.UserRepository, hasher domain.PasswordHasher) *CreateUser {
	return &CreateUser{users: users, hasher: hasher}
}

type CreateUserInput struct {
	OrgID    string
	FullName string
	Email    string
	Password string
	Role     string
}

func (uc *CreateUser) Execute(ctx context.Context, in CreateUserInput) (domain.User, error) {
	role, err := domain.ParseRole(in.Role)
	if err != nil {
		return domain.User{}, err
	}

	hash, err := uc.hasher.Hash(in.Password)
	if err != nil {
		return domain.User{}, err
	}

	user, err := domain.NewUser(NewID(), in.OrgID, in.FullName, in.Email, hash, role)
	if err != nil {
		return domain.User{}, err
	}

	created, err := uc.users.Create(ctx, user)
	if err != nil {
		return domain.User{}, err
	}
	return created, nil
}

// ListUsers returns all users of an organization.
type ListUsers struct {
	users domain.UserRepository
}

func NewListUsers(users domain.UserRepository) *ListUsers {
	return &ListUsers{users: users}
}

func (uc *ListUsers) Execute(ctx context.Context, orgID string) ([]domain.User, error) {
	return uc.users.ListByOrg(ctx, orgID)
}

// GetUser returns a single user by id (used by /auth/me).
type GetUser struct {
	users domain.UserRepository
}

func NewGetUser(users domain.UserRepository) *GetUser {
	return &GetUser{users: users}
}

func (uc *GetUser) Execute(ctx context.Context, id string) (domain.User, error) {
	return uc.users.GetByID(ctx, id)
}
