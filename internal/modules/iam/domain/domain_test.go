package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRole(t *testing.T) {
	cases := []struct {
		in      string
		want    Role
		wantErr bool
	}{
		{"owner", RoleOwner, false},
		{"OWNER", RoleOwner, false},
		{"  manager  ", RoleManager, false},
		{"manager", RoleManager, false},
		{"admin", "", true}, // no admin in Amana (CHANGELOG #2)
		{"", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseRole(tc.in)
			if tc.wantErr {
				require.ErrorIs(t, err, ErrInvalidRole)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNewOrganization(t *testing.T) {
	t.Run("valid with default currency", func(t *testing.T) {
		org, err := NewOrganization("org-1", "  Грозный Мебель  ", "")
		require.NoError(t, err)
		assert.Equal(t, "org-1", org.ID())
		assert.Equal(t, "Грозный Мебель", org.Name())
		assert.Equal(t, "RUB", org.Currency())
		assert.False(t, org.CreatedAt().IsZero())
	})

	t.Run("explicit currency", func(t *testing.T) {
		org, err := NewOrganization("org-1", "Shop", "USD")
		require.NoError(t, err)
		assert.Equal(t, "USD", org.Currency())
	})

	cases := []struct {
		name    string
		id      string
		orgName string
		wantErr error
	}{
		{"empty id", "", "Shop", ErrOrgIDRequired},
		{"empty name", "org-1", "   ", ErrOrgNameRequired},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewOrganization(tc.id, tc.orgName, "RUB")
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestNewUser(t *testing.T) {
	t.Run("valid normalizes email", func(t *testing.T) {
		u, err := NewUser("u1", "o1", "  Иван Петров ", "  Ivan@Example.COM ", "hash", RoleManager)
		require.NoError(t, err)
		assert.Equal(t, "Иван Петров", u.FullName())
		assert.Equal(t, "ivan@example.com", u.Email())
		assert.Equal(t, RoleManager, u.Role())
		assert.Equal(t, "hash", u.PasswordHash())
	})

	cases := []struct {
		name    string
		id      string
		orgID   string
		full    string
		email   string
		hash    string
		role    Role
		wantErr error
	}{
		{"empty id", "", "o1", "Ivan", "i@e.com", "h", RoleOwner, ErrUserIDRequired},
		{"empty org", "u1", "", "Ivan", "i@e.com", "h", RoleOwner, ErrOrgIDRequired},
		{"empty name", "u1", "o1", " ", "i@e.com", "h", RoleOwner, ErrFullNameRequired},
		{"bad email", "u1", "o1", "Ivan", "not-an-email", "h", RoleOwner, ErrInvalidEmail},
		{"empty hash", "u1", "o1", "Ivan", "i@e.com", "", RoleOwner, ErrPasswordHashEmpty},
		{"invalid role", "u1", "o1", "Ivan", "i@e.com", "h", Role("admin"), ErrInvalidRole},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewUser(tc.id, tc.orgID, tc.full, tc.email, tc.hash, tc.role)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestNewApiKey(t *testing.T) {
	t.Run("valid is active", func(t *testing.T) {
		k, err := NewApiKey("k1", "o1", "Marketplace", "amana_abcd1234", HashAPIKey("amana_secret"))
		require.NoError(t, err)
		assert.True(t, k.IsActive())
		assert.Nil(t, k.RevokedAt())
		assert.Equal(t, "Marketplace", k.Name())
	})

	cases := []struct {
		name    string
		id      string
		orgID   string
		keyName string
		prefix  string
		hash    string
		wantErr error
	}{
		{"empty id", "", "o1", "n", "p", "h", ErrUserIDRequired},
		{"empty org", "k1", "", "n", "p", "h", ErrOrgIDRequired},
		{"empty name", "k1", "o1", " ", "p", "h", ErrApiKeyNameRequired},
		{"empty prefix", "k1", "o1", "n", "", "h", ErrApiKeyNameRequired},
		{"empty hash", "k1", "o1", "n", "p", "", ErrApiKeyNameRequired},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewApiKey(tc.id, tc.orgID, tc.keyName, tc.prefix, tc.hash)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestHashAPIKey(t *testing.T) {
	h1 := HashAPIKey("amana_secret")
	h2 := HashAPIKey("amana_secret")
	h3 := HashAPIKey("amana_other")

	assert.Equal(t, h1, h2, "deterministic for same input")
	assert.NotEqual(t, h1, h3, "different input → different hash")
	assert.Len(t, h1, 64, "sha256 hex is 64 chars")
}
