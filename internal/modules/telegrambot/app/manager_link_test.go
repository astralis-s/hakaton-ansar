package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeBotIdentity struct {
	username string
	err      error
}

func (f fakeBotIdentity) Username(context.Context) (string, error) { return f.username, f.err }

func TestEncodeManagerPayload(t *testing.T) {
	require.Equal(t,
		"11111111222233334444555555555555",
		EncodeManagerPayload("11111111-2222-3333-4444-555555555555"))
}

func TestManagerLink(t *testing.T) {
	const managerID = "11111111-2222-3333-4444-555555555555"

	t.Run("builds deep link with hyphen-free payload", func(t *testing.T) {
		uc := NewManagerLink(fakeBotIdentity{username: "amana_support_bot"})
		url, err := uc.Execute(context.Background(), managerID)
		require.NoError(t, err)
		require.Equal(t, "https://t.me/amana_support_bot?start=11111111222233334444555555555555", url)
	})

	t.Run("empty username yields no link", func(t *testing.T) {
		uc := NewManagerLink(fakeBotIdentity{username: ""})
		url, err := uc.Execute(context.Background(), managerID)
		require.NoError(t, err)
		require.Empty(t, url)
	})
}
