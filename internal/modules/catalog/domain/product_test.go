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
		p, err := NewProduct("p1", "o1", "  Диван  ", "Мебель", mustMoney(t, "85000.00"), HalalStatusHalal, 5)
		require.NoError(t, err)
		assert.Equal(t, "Диван", p.Name())
		assert.Equal(t, "Мебель", p.Category())
		assert.Equal(t, "85000.00", p.CostPrice().String())
		assert.Equal(t, 5, p.Stock())
		assert.True(t, p.InStock())
		assert.True(t, p.CanBeFinanced())
		assert.False(t, p.IsHaram())
	})

	t.Run("haram cannot be financed", func(t *testing.T) {
		p, err := NewProduct("p1", "o1", "Вино", "Напитки", mustMoney(t, "1000.00"), HalalStatusHaram, 10)
		require.NoError(t, err)
		assert.True(t, p.IsHaram())
		assert.False(t, p.CanBeFinanced())
	})

	t.Run("doubtful in stock can be financed", func(t *testing.T) {
		p, err := NewProduct("p1", "o1", "Товар", "", mustMoney(t, "500.00"), HalalStatusDoubtful, 1)
		require.NoError(t, err)
		assert.True(t, p.CanBeFinanced())
	})

	t.Run("out of stock cannot be financed", func(t *testing.T) {
		p, err := NewProduct("p1", "o1", "Товар", "", mustMoney(t, "500.00"), HalalStatusHalal, 0)
		require.NoError(t, err)
		assert.False(t, p.InStock())
		assert.False(t, p.CanBeFinanced())
	})

	cases := []struct {
		name    string
		id      string
		orgID   string
		pname   string
		cost    string
		status  HalalStatus
		stock   int
		wantErr error
	}{
		{"empty id", "", "o1", "X", "100.00", HalalStatusHalal, 1, ErrProductIDRequired},
		{"empty org", "p1", "", "X", "100.00", HalalStatusHalal, 1, ErrOrgIDRequired},
		{"empty name", "p1", "o1", "  ", "100.00", HalalStatusHalal, 1, ErrProductNameRequired},
		{"zero cost", "p1", "o1", "X", "0.00", HalalStatusHalal, 1, ErrCostPriceNotPositive},
		{"invalid status", "p1", "o1", "X", "100.00", HalalStatus("x"), 1, ErrInvalidHalalStatus},
		{"negative stock", "p1", "o1", "X", "100.00", HalalStatusHalal, -1, ErrNegativeStock},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewProduct(tc.id, tc.orgID, tc.pname, "", mustMoney(t, tc.cost), tc.status, tc.stock)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestProductWithStockDelta(t *testing.T) {
	p, err := NewProduct("p1", "o1", "Товар", "", mustMoney(t, "100.00"), HalalStatusHalal, 3)
	require.NoError(t, err)

	t.Run("receipt increases stock", func(t *testing.T) {
		up, err := p.WithStockDelta(2)
		require.NoError(t, err)
		assert.Equal(t, 5, up.Stock())
		assert.Equal(t, 3, p.Stock()) // original unchanged (value semantics)
	})

	t.Run("sale to exactly zero is allowed", func(t *testing.T) {
		up, err := p.WithStockDelta(-3)
		require.NoError(t, err)
		assert.Equal(t, 0, up.Stock())
		assert.False(t, up.InStock())
	})

	t.Run("oversell is rejected", func(t *testing.T) {
		_, err := p.WithStockDelta(-4)
		require.ErrorIs(t, err, ErrInsufficientStock)
	})
}

func TestProductUpdatePreservesIdentityAndStock(t *testing.T) {
	p, err := NewProduct("p1", "o1", "Old", "Cat", mustMoney(t, "100.00"), HalalStatusHalal, 7)
	require.NoError(t, err)

	updated, err := p.Update("New", "Cat2", mustMoney(t, "200.00"), HalalStatusDoubtful)
	require.NoError(t, err)
	assert.Equal(t, "p1", updated.ID())
	assert.Equal(t, p.CreatedAt(), updated.CreatedAt())
	assert.Equal(t, "New", updated.Name())
	assert.Equal(t, "200.00", updated.CostPrice().String())
	assert.Equal(t, 7, updated.Stock()) // stock preserved across edits

	_, err = p.Update("New", "Cat2", mustMoney(t, "0.00"), HalalStatusHalal)
	require.ErrorIs(t, err, ErrCostPriceNotPositive)
}
