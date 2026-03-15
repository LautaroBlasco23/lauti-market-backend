package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/order/domain"
)

func TestOrderStatus_IsValid_AllStatuses(t *testing.T) {
	statuses := []domain.OrderStatus{
		domain.StatusPending,
		domain.StatusConfirmed,
		domain.StatusShipped,
		domain.StatusDelivered,
		domain.StatusCancelled,
	}
	for _, s := range statuses {
		assert.True(t, s.IsValid(), "expected %q to be valid", s)
	}
}

func TestOrderStatus_IsValid_Invalid(t *testing.T) {
	assert.False(t, domain.OrderStatus("unknown").IsValid())
	assert.False(t, domain.OrderStatus("").IsValid())
}
