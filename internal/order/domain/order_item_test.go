package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/order/domain"
)

func TestNewOrderItem_Valid(t *testing.T) {
	item, err := domain.NewOrderItem("item-1", "order-1", "product-1", 2, 9.99)
	require.NoError(t, err)
	assert.Equal(t, "item-1", item.ID())
	assert.Equal(t, "order-1", item.OrderID())
	assert.Equal(t, "product-1", item.ProductID())
	assert.Equal(t, 2, item.Quantity())
	assert.Equal(t, 9.99, item.UnitPrice())
}

func TestNewOrderItem_ZeroQuantity(t *testing.T) {
	_, err := domain.NewOrderItem("item-1", "order-1", "product-1", 0, 9.99)
	assert.ErrorIs(t, err, apiDomain.ErrInvalidQuantity)
}

func TestNewOrderItem_NegativeQuantity(t *testing.T) {
	_, err := domain.NewOrderItem("item-1", "order-1", "product-1", -1, 9.99)
	assert.ErrorIs(t, err, apiDomain.ErrInvalidQuantity)
}

func TestNewOrderItem_ZeroPrice(t *testing.T) {
	_, err := domain.NewOrderItem("item-1", "order-1", "product-1", 1, 0)
	assert.ErrorIs(t, err, apiDomain.ErrInvalidPrice)
}

func TestNewOrderItem_NegativePrice(t *testing.T) {
	_, err := domain.NewOrderItem("item-1", "order-1", "product-1", 1, -5.0)
	assert.ErrorIs(t, err, apiDomain.ErrInvalidPrice)
}
