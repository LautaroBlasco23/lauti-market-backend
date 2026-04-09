package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/payment/domain"
)

func TestNewPayment_Defaults(t *testing.T) {
	before := time.Now()
	p := domain.NewPayment("pay-1", "order-1", "user-1", "pref-1", 100.0)
	after := time.Now()

	assert.Equal(t, "pay-1", p.ID())
	assert.Equal(t, "order-1", p.OrderID())
	assert.Equal(t, "user-1", p.UserID())
	assert.Equal(t, "pref-1", p.PreferenceID())
	assert.Equal(t, 100.0, p.Amount())
	assert.Equal(t, "ARS", p.Currency())
	assert.Equal(t, domain.StatusPending, p.Status())
	assert.Equal(t, int64(0), p.MPPaymentID())
	assert.Empty(t, p.StatusDetail())
	assert.Empty(t, p.PaymentMethod())
	assert.True(t, !p.CreatedAt().Before(before))
	assert.True(t, !p.CreatedAt().After(after))
}

func TestNewPaymentFromDB_Fields(t *testing.T) {
	created := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	updated := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	p := domain.NewPaymentFromDB(
		"pay-1", "order-1", "user-1",
		987654, 200.0, "ARS",
		domain.StatusApproved,
		"accredited", "credit_card", "pref-abc",
		created, updated,
	)

	assert.Equal(t, "pay-1", p.ID())
	assert.Equal(t, "order-1", p.OrderID())
	assert.Equal(t, "user-1", p.UserID())
	assert.Equal(t, int64(987654), p.MPPaymentID())
	assert.Equal(t, 200.0, p.Amount())
	assert.Equal(t, "ARS", p.Currency())
	assert.Equal(t, domain.StatusApproved, p.Status())
	assert.Equal(t, "accredited", p.StatusDetail())
	assert.Equal(t, "credit_card", p.PaymentMethod())
	assert.Equal(t, "pref-abc", p.PreferenceID())
	assert.Equal(t, created, p.CreatedAt())
	assert.Equal(t, updated, p.UpdatedAt())
}

func TestUpdateFromMP(t *testing.T) {
	p := domain.NewPayment("pay-1", "order-1", "user-1", "pref-1", 100.0)
	before := time.Now()

	p.UpdateFromMP(555, domain.StatusApproved, "accredited", "debit_card")

	assert.Equal(t, int64(555), p.MPPaymentID())
	assert.Equal(t, domain.StatusApproved, p.Status())
	assert.Equal(t, "accredited", p.StatusDetail())
	assert.Equal(t, "debit_card", p.PaymentMethod())
	assert.True(t, !p.UpdatedAt().Before(before))
}
