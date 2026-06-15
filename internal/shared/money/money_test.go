package money

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromStringAndString(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"120000", "120000.00", false},
		{"120000.5", "120000.50", false},
		{"0", "0.00", false},
		{"85000.99", "85000.99", false},
		{"  1500.00 ", "1500.00", false},
		{"abc", "", true},
		{"", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			m, err := FromString(tc.in, "RUB")
			if tc.wantErr {
				require.ErrorIs(t, err, ErrInvalidAmount)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, m.String())
			assert.Equal(t, "RUB", m.Currency())
		})
	}
}

func TestCents(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"120000.00", 12000000},
		{"0.01", 1},
		{"33333.34", 3333334},
		{"0.00", 0},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			m, err := FromString(tc.in, "RUB")
			require.NoError(t, err)
			assert.Equal(t, tc.want, m.Cents())
		})
	}
}

func TestFromCentsRoundTrip(t *testing.T) {
	m := FromCents(3333334, "RUB")
	assert.Equal(t, "33333.34", m.String())
	assert.Equal(t, int64(3333334), m.Cents())
}

func TestAddSub(t *testing.T) {
	a, _ := FromString("100.00", "RUB")
	b, _ := FromString("33.33", "RUB")

	sum, err := a.Add(b)
	require.NoError(t, err)
	assert.Equal(t, "133.33", sum.String())

	diff, err := a.Sub(b)
	require.NoError(t, err)
	assert.Equal(t, "66.67", diff.String())
}

func TestCurrencyMismatch(t *testing.T) {
	a, _ := FromString("100.00", "RUB")
	b, _ := FromString("100.00", "USD")

	_, err := a.Add(b)
	require.ErrorIs(t, err, ErrCurrencyMismatch)
	_, err = a.Sub(b)
	require.ErrorIs(t, err, ErrCurrencyMismatch)
	_, err = a.Cmp(b)
	require.ErrorIs(t, err, ErrCurrencyMismatch)
}

func TestSigns(t *testing.T) {
	pos, _ := FromString("1.00", "RUB")
	zero := Zero("RUB")
	neg := New(decimal.RequireFromString("-5.00"), "RUB")

	assert.True(t, pos.IsPositive())
	assert.False(t, pos.IsZero())
	assert.True(t, zero.IsZero())
	assert.True(t, neg.IsNegative())
}

func TestRoundsToTwoPlaces(t *testing.T) {
	m := New(decimal.RequireFromString("10.005"), "RUB")
	// banker-independent: decimal.Round(2) rounds half away from zero → 10.01
	assert.Equal(t, "10.01", m.String())
}
