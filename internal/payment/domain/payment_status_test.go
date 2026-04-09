package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/payment/domain"
)

func TestPaymentStatus_IsValid(t *testing.T) {
	valid := []domain.PaymentStatus{
		domain.StatusPending,
		domain.StatusApproved,
		domain.StatusRejected,
		domain.StatusCancelled,
		domain.StatusInProcess,
	}
	for _, s := range valid {
		assert.True(t, s.IsValid(), "expected %q to be valid", s)
	}

	assert.False(t, domain.PaymentStatus("unknown").IsValid())
	assert.False(t, domain.PaymentStatus("").IsValid())
}

func TestPaymentStatus_IsTerminal(t *testing.T) {
	terminal := []domain.PaymentStatus{
		domain.StatusApproved,
		domain.StatusRejected,
		domain.StatusCancelled,
	}
	for _, s := range terminal {
		assert.True(t, s.IsTerminal(), "expected %q to be terminal", s)
	}

	nonTerminal := []domain.PaymentStatus{
		domain.StatusPending,
		domain.StatusInProcess,
	}
	for _, s := range nonTerminal {
		assert.False(t, s.IsTerminal(), "expected %q to be non-terminal", s)
	}
}
