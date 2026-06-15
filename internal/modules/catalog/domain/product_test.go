package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

func mustMoney(t *testing.T, s string) money.Money {
	t.Helper()
	m, err := money.FromString(s, money.DefaultCurrency)
	require.NoError(t, err)
	return m
}

func TestParseHalalStatus(t *testing.T) {
	cases := []struct {
		in      string
		want    HalalStatus
		wantErr bool
	}{
		{"halal", HalalStatusHalal, false},
		{"HARAM", HalalStatusHaram, false},
		{" doubtful ", HalalStatusDoubtful, false},
		{"mushbooh", "", true},
		{"", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseHalalStatus(tc.in)
			if tc.wantErr {
				require.ErrorIs(t, err, ErrInvalidHalalStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNewProduct(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		p, err := NewProduct("p1", "o1", "  Диван  ", "Мебель", mustMoney(t, "85000.00"), HalalStatusHalal)
		require.NoError(t, err)
		assert.Equal(t, "Диван", p.Name())
		assert.Equal(t, "Мебель", p.Category())
		assert.Equal(t, "85000.00", p.CostPrice().String())
		assert.True(t, p.CanBeFinanced())
		assert.False(t, p.IsHaram())
	})

	t.Run("haram cannot be financed", func(t *testing.T) {
		p, err := NewProduct("p1", "o1", "Вино", "Напитки", mustMoney(t, "1000.00"), HalalStatusHaram)
		require.NoError(t, err)
		assert.True(t, p.IsHaram())
		assert.False(t, p.CanBeFinanced())
	})

	t.Run("doubtful can be financed", func(t *testing.T) {
		p, err := NewProduct("p1", "o1", "Товар", "", mustMoney(t, "500.00"), HalalStatusDoubtful)
		require.NoError(t, err)
		assert.True(t, p.CanBeFinanced())
	})

	cases := []struct {
		name    string
		id      string
		orgID   string
		pname   string
		cost    string
		status  HalalStatus
		wantErr error
	}{
		{"empty id", "", "o1", "X", "100.00", HalalStatusHalal, ErrProductIDRequired},
		{"empty org", "p1", "", "X", "100.00", HalalStatusHalal, ErrOrgIDRequired},
		{"empty name", "p1", "o1", "  ", "100.00", HalalStatusHalal, ErrProductNameRequired},
		{"zero cost", "p1", "o1", "X", "0.00", HalalStatusHalal, ErrCostPriceNotPositive},
		{"invalid status", "p1", "o1", "X", "100.00", HalalStatus("x"), ErrInvalidHalalStatus},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewProduct(tc.id, tc.orgID, tc.pname, "", mustMoney(t, tc.cost), tc.status)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestProductUpdatePreservesIdentity(t *testing.T) {
	p, err := NewProduct("p1", "o1", "Old", "Cat", mustMoney(t, "100.00"), HalalStatusHalal)
	require.NoError(t, err)

	updated, err := p.Update("New", "Cat2", mustMoney(t, "200.00"), HalalStatusDoubtful)
	require.NoError(t, err)
	assert.Equal(t, "p1", updated.ID())
	assert.Equal(t, p.CreatedAt(), updated.CreatedAt())
	assert.Equal(t, "New", updated.Name())
	assert.Equal(t, "200.00", updated.CostPrice().String())

	_, err = p.Update("New", "Cat2", mustMoney(t, "0.00"), HalalStatusHalal)
	require.ErrorIs(t, err, ErrCostPriceNotPositive)
}
