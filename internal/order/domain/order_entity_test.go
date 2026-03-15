package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/order/domain"
)

func makeItem(t *testing.T) *domain.OrderItem {
	t.Helper()
	item, err := domain.NewOrderItem("item-1", "order-1", "product-1", 1, 10.0)
	require.NoError(t, err)
	return item
}

func TestNewOrder_Valid(t *testing.T) {
	items := []*domain.OrderItem{makeItem(t)}
	order, err := domain.NewOrder("order-1", "user-1", "store-1", items, 10.0)
	require.NoError(t, err)
	assert.Equal(t, "order-1", order.ID())
	assert.Equal(t, "user-1", order.UserID())
	assert.Equal(t, "store-1", order.StoreID())
	assert.Equal(t, domain.StatusPending, order.Status())
	assert.Equal(t, 10.0, order.TotalPrice())
	assert.Len(t, order.Items(), 1)
}

func TestNewOrder_EmptyItems(t *testing.T) {
	_, err := domain.NewOrder("order-1", "user-1", "store-1", []*domain.OrderItem{}, 10.0)
	assert.ErrorIs(t, err, apiDomain.ErrEmptyOrderItems)
}

func TestNewOrder_ZeroTotalPrice(t *testing.T) {
	items := []*domain.OrderItem{makeItem(t)}
	_, err := domain.NewOrder("order-1", "user-1", "store-1", items, 0)
	assert.ErrorIs(t, err, apiDomain.ErrInvalidPrice)
}

// TransitionStatus tests

func TestTransition_PendingToConfirmed_StoreOwner(t *testing.T) {
	items := []*domain.OrderItem{makeItem(t)}
	order, err := domain.NewOrder("order-1", "user-1", "store-1", items, 10.0)
	require.NoError(t, err)

	err = order.TransitionStatus(domain.StatusConfirmed, "store", "store-1")
	assert.NoError(t, err)
	assert.Equal(t, domain.StatusConfirmed, order.Status())
}

func TestTransition_PendingToConfirmed_WrongStore(t *testing.T) {
	items := []*domain.OrderItem{makeItem(t)}
	order, err := domain.NewOrder("order-1", "user-1", "store-1", items, 10.0)
	require.NoError(t, err)

	err = order.TransitionStatus(domain.StatusConfirmed, "store", "other-store")
	assert.ErrorIs(t, err, apiDomain.ErrForbidden)
}

func TestTransition_PendingToConfirmed_UserAccount(t *testing.T) {
	items := []*domain.OrderItem{makeItem(t)}
	order, err := domain.NewOrder("order-1", "user-1", "store-1", items, 10.0)
	require.NoError(t, err)

	err = order.TransitionStatus(domain.StatusConfirmed, "user", "user-1")
	assert.ErrorIs(t, err, apiDomain.ErrForbidden)
}

func TestTransition_ConfirmedToShipped_StoreOwner(t *testing.T) {
	items := []*domain.OrderItem{makeItem(t)}
	order, err := domain.NewOrder("order-1", "user-1", "store-1", items, 10.0)
	require.NoError(t, err)
	require.NoError(t, order.TransitionStatus(domain.StatusConfirmed, "store", "store-1"))

	err = order.TransitionStatus(domain.StatusShipped, "store", "store-1")
	assert.NoError(t, err)
	assert.Equal(t, domain.StatusShipped, order.Status())
}

func TestTransition_ShippedToDelivered_StoreOwner(t *testing.T) {
	items := []*domain.OrderItem{makeItem(t)}
	order, err := domain.NewOrder("order-1", "user-1", "store-1", items, 10.0)
	require.NoError(t, err)
	require.NoError(t, order.TransitionStatus(domain.StatusConfirmed, "store", "store-1"))
	require.NoError(t, order.TransitionStatus(domain.StatusShipped, "store", "store-1"))

	err = order.TransitionStatus(domain.StatusDelivered, "store", "store-1")
	assert.NoError(t, err)
	assert.Equal(t, domain.StatusDelivered, order.Status())
}

func TestTransition_PendingToCancelled_UserOwner(t *testing.T) {
	items := []*domain.OrderItem{makeItem(t)}
	order, err := domain.NewOrder("order-1", "user-1", "store-1", items, 10.0)
	require.NoError(t, err)

	err = order.TransitionStatus(domain.StatusCancelled, "user", "user-1")
	assert.NoError(t, err)
	assert.Equal(t, domain.StatusCancelled, order.Status())
}

func TestTransition_PendingToCancelled_WrongUser(t *testing.T) {
	items := []*domain.OrderItem{makeItem(t)}
	order, err := domain.NewOrder("order-1", "user-1", "store-1", items, 10.0)
	require.NoError(t, err)

	err = order.TransitionStatus(domain.StatusCancelled, "user", "other-user")
	assert.ErrorIs(t, err, apiDomain.ErrForbidden)
}

func TestTransition_PendingToCancelled_StoreAccount(t *testing.T) {
	items := []*domain.OrderItem{makeItem(t)}
	order, err := domain.NewOrder("order-1", "user-1", "store-1", items, 10.0)
	require.NoError(t, err)

	err = order.TransitionStatus(domain.StatusCancelled, "store", "store-1")
	assert.ErrorIs(t, err, apiDomain.ErrForbidden)
}

func TestTransition_InvalidTransition(t *testing.T) {
	items := []*domain.OrderItem{makeItem(t)}
	order, err := domain.NewOrder("order-1", "user-1", "store-1", items, 10.0)
	require.NoError(t, err)

	// Pending → Shipped is not a valid transition
	err = order.TransitionStatus(domain.StatusShipped, "store", "store-1")
	assert.ErrorIs(t, err, apiDomain.ErrForbiddenTransition)
}
