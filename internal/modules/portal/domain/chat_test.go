package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMessage(t *testing.T) {
	t.Run("valid trims body", func(t *testing.T) {
		m, err := NewMessage("m1", "c1", "o1", SenderClient, "cl1", "  привет  ")
		require.NoError(t, err)
		assert.Equal(t, "привет", m.Body())
		assert.Equal(t, SenderClient, m.SenderKind())
	})

	cases := []struct {
		name    string
		id      string
		convID  string
		orgID   string
		kind    SenderKind
		sender  string
		body    string
		wantErr error
	}{
		{"empty id", "", "c1", "o1", SenderStaff, "u1", "hi", ErrMessageIDRequired},
		{"empty conversation", "m1", "", "o1", SenderStaff, "u1", "hi", ErrConversationIDRequired},
		{"empty org", "m1", "c1", "", SenderStaff, "u1", "hi", ErrOrgIDRequired},
		{"invalid kind", "m1", "c1", "o1", SenderKind("bot"), "u1", "hi", ErrInvalidSenderKind},
		{"empty sender", "m1", "c1", "o1", SenderStaff, "", "hi", ErrSenderIDRequired},
		{"blank body", "m1", "c1", "o1", SenderStaff, "u1", "   ", ErrMessageBodyRequired},
		{"too long", "m1", "c1", "o1", SenderStaff, "u1", strings.Repeat("x", maxMessageLen+1), ErrMessageTooLong},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewMessage(tc.id, tc.convID, tc.orgID, tc.kind, tc.sender, tc.body)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestNewPortalAccount(t *testing.T) {
	t.Run("valid normalizes email", func(t *testing.T) {
		a, err := NewPortalAccount("c1", "o1", "  Client@Mail.RU ", "hash")
		require.NoError(t, err)
		assert.Equal(t, "client@mail.ru", a.Email())
	})

	cases := []struct {
		name     string
		clientID string
		orgID    string
		email    string
		hash     string
		wantErr  error
	}{
		{"empty client", "", "o1", "a@b.ru", "h", ErrClientIDRequired},
		{"empty org", "c1", "", "a@b.ru", "h", ErrOrgIDRequired},
		{"bad email", "c1", "o1", "notanemail", "h", ErrInvalidEmail},
		{"empty hash", "c1", "o1", "a@b.ru", "", ErrPasswordRequired},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewPortalAccount(tc.clientID, tc.orgID, tc.email, tc.hash)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestSenderKindValid(t *testing.T) {
	assert.True(t, SenderClient.Valid())
	assert.True(t, SenderStaff.Valid())
	assert.False(t, SenderKind("x").Valid())
}
