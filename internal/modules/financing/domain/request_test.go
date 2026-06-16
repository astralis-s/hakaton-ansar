package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

func rubReq(t *testing.T, s string) money.Money {
	t.Helper()
	m, err := money.FromString(s, "RUB")
	require.NoError(t, err)
	return m
}

func TestNewContractRequest(t *testing.T) {
	t.Run("valid is pending and trims note", func(t *testing.T) {
		r, err := NewContractRequest("r1", "o1", "c1", "p1", 6, rubReq(t, "10000.00"), "  хочу диван  ")
		require.NoError(t, err)
		assert.Equal(t, RequestPending, r.Status())
		assert.Equal(t, "хочу диван", r.Note())
		assert.Equal(t, "", r.ContractID())
		assert.Nil(t, r.DecidedAt())
	})

	cases := []struct {
		name         string
		id, org      string
		client, prod string
		installments int
		down         string
		wantErr      error
	}{
		{"empty id", "", "o1", "c1", "p1", 6, "0.00", ErrRequestIDRequired},
		{"empty org", "r1", "", "c1", "p1", 6, "0.00", ErrOrgIDRequired},
		{"empty client", "r1", "o1", "", "p1", 6, "0.00", ErrClientIDRequired},
		{"empty product", "r1", "o1", "c1", "", 6, "0.00", ErrProductIDRequired},
		{"zero installments", "r1", "o1", "c1", "p1", 0, "0.00", ErrDesiredInstallmentsInvalid},
		{"negative down", "r1", "o1", "c1", "p1", 6, "-1.00", ErrDownPaymentNegative},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewContractRequest(tc.id, tc.org, tc.client, tc.prod, tc.installments, rubReq(t, tc.down), "")
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestContractRequestApprove(t *testing.T) {
	t.Run("approve links contract and stamps decided", func(t *testing.T) {
		r, _ := NewContractRequest("r1", "o1", "c1", "p1", 6, rubReq(t, "0.00"), "")
		require.NoError(t, r.Approve("contract-123"))
		assert.Equal(t, RequestApproved, r.Status())
		assert.Equal(t, "contract-123", r.ContractID())
		require.NotNil(t, r.DecidedAt())
	})

	t.Run("approve requires a contract id", func(t *testing.T) {
		r, _ := NewContractRequest("r1", "o1", "c1", "p1", 6, rubReq(t, "0.00"), "")
		require.ErrorIs(t, r.Approve(""), ErrContractIDRequired)
	})

	t.Run("cannot approve a non-pending request", func(t *testing.T) {
		r, _ := NewContractRequest("r1", "o1", "c1", "p1", 6, rubReq(t, "0.00"), "")
		require.NoError(t, r.Reject())
		require.ErrorIs(t, r.Approve("c1"), ErrRequestNotPending)
	})
}

func TestContractRequestReject(t *testing.T) {
	r, _ := NewContractRequest("r1", "o1", "c1", "p1", 6, rubReq(t, "0.00"), "")
	require.NoError(t, r.Reject())
	assert.Equal(t, RequestRejected, r.Status())
	require.NotNil(t, r.DecidedAt())
	// rejecting again is rejected
	require.ErrorIs(t, r.Reject(), ErrRequestNotPending)
}
