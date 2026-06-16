package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStockReason(t *testing.T) {
	cases := []struct {
		in      string
		want    StockReason
		wantErr bool
	}{
		{"receipt", StockReceipt, false},
		{"SALE", StockSale, false},
		{" adjustment ", StockAdjustment, false},
		{"writeoff", StockWriteoff, false},
		{"gift", "", true},
		{"", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseStockReason(tc.in)
			if tc.wantErr {
				require.ErrorIs(t, err, ErrInvalidStockReason)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.True(t, got.Valid())
		})
	}
}

func TestNewStockMovement(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		m, err := NewStockMovement("m1", "o1", "p1", -1, StockSale, "  по договору  ", 4)
		require.NoError(t, err)
		assert.Equal(t, -1, m.Delta())
		assert.Equal(t, StockSale, m.Reason())
		assert.Equal(t, "по договору", m.Note()) // trimmed
		assert.Equal(t, 4, m.BalanceAfter())
	})

	cases := []struct {
		name      string
		id        string
		orgID     string
		productID string
		delta     int
		reason    StockReason
		wantErr   error
	}{
		{"empty id", "", "o1", "p1", 1, StockReceipt, ErrProductIDRequired},
		{"empty org", "m1", "", "p1", 1, StockReceipt, ErrProductIDRequired},
		{"empty product", "m1", "o1", "", 1, StockReceipt, ErrProductIDRequired},
		{"zero delta", "m1", "o1", "p1", 0, StockReceipt, ErrStockDeltaZero},
		{"invalid reason", "m1", "o1", "p1", 1, StockReason("x"), ErrInvalidStockReason},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewStockMovement(tc.id, tc.orgID, tc.productID, tc.delta, tc.reason, "", 0)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}
