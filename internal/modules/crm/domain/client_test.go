package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("valid trims fields", func(t *testing.T) {
		c, err := NewClient("c1", "o1", "  Магомед Алиев ", " +79280000000 ", " 96 00 123456 ")
		require.NoError(t, err)
		assert.Equal(t, "Магомед Алиев", c.FullName())
		assert.Equal(t, "+79280000000", c.Phone())
		assert.Equal(t, "96 00 123456", c.Document())
	})

	t.Run("optional fields may be empty", func(t *testing.T) {
		c, err := NewClient("c1", "o1", "Иса", "", "")
		require.NoError(t, err)
		assert.Empty(t, c.Phone())
		assert.Empty(t, c.Document())
	})

	cases := []struct {
		name    string
		id      string
		orgID   string
		full    string
		wantErr error
	}{
		{"empty id", "", "o1", "Name", ErrClientIDRequired},
		{"empty org", "c1", "", "Name", ErrOrgIDRequired},
		{"empty name", "c1", "o1", "   ", ErrClientNameRequired},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewClient(tc.id, tc.orgID, tc.full, "", "")
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestClientUpdatePreservesIdentity(t *testing.T) {
	c, err := NewClient("c1", "o1", "Old Name", "", "")
	require.NoError(t, err)

	updated, err := c.Update("New Name", "+79281112233", "doc")
	require.NoError(t, err)
	assert.Equal(t, "c1", updated.ID())
	assert.Equal(t, c.CreatedAt(), updated.CreatedAt())
	assert.Equal(t, "New Name", updated.FullName())

	_, err = c.Update("  ", "", "")
	require.ErrorIs(t, err, ErrClientNameRequired)
}
