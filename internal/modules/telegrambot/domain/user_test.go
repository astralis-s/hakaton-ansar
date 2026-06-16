package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
		ok   bool
	}{
		{"ru +7 с разделителями", "+7 928 000-00-00", "+79280000000", true},
		{"ru 8 в скобках", "8 (928) 000 00 00", "+79280000000", true},
		{"ru 11 цифр с 7", "79280000000", "+79280000000", true},
		{"ru 10 цифр без кода", "9280000000", "+79280000000", true},
		{"межд. с плюсом", "+1 202 555 0143", "+12025550143", true},
		{"слишком коротко", "12345", "", false},
		{"пусто", "", "", false},
		{"буквы", "не телефон", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := NormalizePhone(tt.in)
			require.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidFullName(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"Ибрагимов Ислам Висханович", true},
		{"Иванов Иван", true},
		{"Ан Ли", true},
		{"Иван", false},   // одно слово
		{"/start", false}, // команда
		{"", false},
	}
	for _, tt := range tests {
		assert.Equalf(t, tt.want, IsValidFullName(NormalizeFullName(tt.in)), "name=%q", tt.in)
	}
}

func TestNewTelegramUserValidation(t *testing.T) {
	_, err := NewTelegramUser(0, "org-1", "")
	require.ErrorIs(t, err, ErrChatIDRequired)

	_, err = NewTelegramUser(1, "", "")
	require.ErrorIs(t, err, ErrOrgIDRequired)
}

func TestRegistrationFSM(t *testing.T) {
	u, err := NewTelegramUser(123, "org-1", "ivan")
	require.NoError(t, err)
	require.Equal(t, StateAwaitingName, u.State())

	// Невалидное ФИО не продвигает состояние.
	require.ErrorIs(t, u.RecordName("Иван"), ErrInvalidFullName)
	require.Equal(t, StateAwaitingName, u.State())

	require.NoError(t, u.RecordName("  Ибрагимов   Ислам Висханович "))
	require.Equal(t, StateAwaitingPhone, u.State())
	assert.Equal(t, "Ибрагимов Ислам Висханович", u.FullName())

	// Нельзя завершить регистрацию без клиента.
	require.ErrorIs(t, u.CompleteRegistration("+79280000000", ""), ErrClientIDRequired)
	require.False(t, u.Registered())

	require.NoError(t, u.CompleteRegistration("+79280000000", "client-1"))
	require.True(t, u.Registered())
	assert.Equal(t, "client-1", u.ClientID())
	assert.Equal(t, "+79280000000", u.Phone())

	// Запись ФИО в завершённом состоянии запрещена.
	require.ErrorIs(t, u.RecordName("Пётр Петров"), ErrInvalidState)
}

func TestRestartRegistration(t *testing.T) {
	u, _ := NewTelegramUser(123, "org-1", "")
	require.NoError(t, u.RecordName("Иванов Иван"))
	u.RestartRegistration()
	require.Equal(t, StateAwaitingName, u.State())
	assert.Empty(t, u.FullName())
}
