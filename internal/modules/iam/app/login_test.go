package app

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/astralis-s/hakaton-ansar/internal/modules/iam/domain"
)

// --- fakes ------------------------------------------------------------------

type fakeUserRepo struct {
	user   domain.User
	getErr error
}

func (f *fakeUserRepo) Create(context.Context, domain.User) (domain.User, error) {
	return domain.User{}, nil
}
func (f *fakeUserRepo) GetByEmail(context.Context, string) (domain.User, error) {
	return f.user, f.getErr
}
func (f *fakeUserRepo) GetByID(context.Context, string) (domain.User, error) {
	return f.user, f.getErr
}
func (f *fakeUserRepo) ListByOrg(context.Context, string) ([]domain.User, error) { return nil, nil }
func (f *fakeUserRepo) ExistsByEmail(context.Context, string) (bool, error)      { return false, nil }

type fakeHasher struct{ matchErr error }

func (f *fakeHasher) Hash(string) (string, error) { return "hash", nil }
func (f *fakeHasher) Compare(_, _ string) error   { return f.matchErr }

type fakeTokens struct{ token string }

func (f *fakeTokens) Issue(domain.User) (string, time.Time, error) {
	return f.token, time.Now().Add(time.Hour), nil
}
func (f *fakeTokens) Parse(string) (domain.Principal, error) { return domain.Principal{}, nil }

func validUser(t *testing.T) domain.User {
	t.Helper()
	u, err := domain.NewUser("u1", "o1", "Owner", "owner@example.com", "stored-hash", domain.RoleOwner)
	require.NoError(t, err)
	return u
}

// --- tests ------------------------------------------------------------------

func TestLogin(t *testing.T) {
	t.Run("success issues token", func(t *testing.T) {
		uc := NewLogin(
			&fakeUserRepo{user: validUser(t)},
			&fakeHasher{matchErr: nil},
			&fakeTokens{token: "jwt-123"},
		)
		out, err := uc.Execute(context.Background(), LoginInput{Email: "owner@example.com", Password: "pw"})
		require.NoError(t, err)
		assert.Equal(t, "jwt-123", out.Token)
		assert.Equal(t, "owner@example.com", out.User.Email())
		assert.False(t, out.ExpiresAt.IsZero())
	})

	t.Run("unknown email → invalid credentials (no enumeration)", func(t *testing.T) {
		uc := NewLogin(
			&fakeUserRepo{getErr: domain.ErrUserNotFound},
			&fakeHasher{},
			&fakeTokens{},
		)
		_, err := uc.Execute(context.Background(), LoginInput{Email: "ghost@example.com", Password: "pw"})
		require.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("wrong password → invalid credentials", func(t *testing.T) {
		uc := NewLogin(
			&fakeUserRepo{user: validUser(t)},
			&fakeHasher{matchErr: domain.ErrInvalidCredentials},
			&fakeTokens{},
		)
		_, err := uc.Execute(context.Background(), LoginInput{Email: "owner@example.com", Password: "bad"})
		require.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})
}
