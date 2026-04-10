//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	orderDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/order/domain"
	orderRepo "github.com/LautaroBlasco23/lauti-market-backend/internal/order/infrastructure/repository"
	paymentDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/payment/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/payment/infrastructure/repository"
	productDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/product/domain"
	productRepo "github.com/LautaroBlasco23/lauti-market-backend/internal/product/infrastructure/repository"
	storeDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
	storeRepo "github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure/repository"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/testutil"
	userDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/user/domain"
	userRepo "github.com/LautaroBlasco23/lauti-market-backend/internal/user/infrastructure/repository"
)

// setupDeps seeds all FK rows required before inserting a payment and returns
// the payment repository plus the seeded order.
func setupDeps(t *testing.T) (context.Context, *repository.PaymentPostgresRepository, *orderDomain.Order) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	ctx := context.Background()

	ur := userRepo.NewUserPostgresRepository(db)
	user, _ := userDomain.NewUser("user-1", "John", "Doe")
	require.NoError(t, ur.Save(ctx, user))

	sr := storeRepo.NewStorePostgresRepository(db)
	store, _ := storeDomain.NewStore("store-1", storeDomain.CreateStoreInput{
		Name:        "Test Store",
		Description: "A test store description",
		Address:     "123 St",
		PhoneNumber: "12345678",
	})
	require.NoError(t, sr.Save(ctx, store))

	pr := productRepo.NewProductPostgresRepository(db)
	product, _ := productDomain.NewProduct("prod-1", "store-1", "Widget", "A detailed product description", "electronics", 10, 100.0, nil)
	require.NoError(t, pr.Save(ctx, product))

	or := orderRepo.NewOrderPostgresRepository(db)
	item, _ := orderDomain.NewOrderItem("item-1", "order-1", "prod-1", 1, 100.0)
	order, _ := orderDomain.NewOrder("order-1", "user-1", "store-1", []*orderDomain.OrderItem{item}, 100.0)
	require.NoError(t, or.Save(ctx, order))

	repo := repository.NewPaymentPostgresRepository(db)
	return ctx, repo, order
}

func makePayment(id, orderID, userID string) *paymentDomain.Payment {
	return paymentDomain.NewPayment(id, orderID, userID, "pref-"+id, 100.0)
}

// --- Save / FindByID ---

func TestPaymentRepository_Save_FindByID(t *testing.T) {
	ctx, repo, order := setupDeps(t)

	p := makePayment("pay-1", order.ID(), "user-1")
	require.NoError(t, repo.Save(ctx, p))

	found, err := repo.FindByID(ctx, "pay-1")
	require.NoError(t, err)
	assert.Equal(t, "pay-1", found.ID())
	assert.Equal(t, order.ID(), found.OrderID())
	assert.Equal(t, "user-1", found.UserID())
	assert.Equal(t, paymentDomain.StatusPending, found.Status())
	assert.Equal(t, 100.0, found.Amount())
	assert.Equal(t, "ARS", found.Currency())
}

func TestPaymentRepository_FindByID_NotFound(t *testing.T) {
	ctx, repo, _ := setupDeps(t)

	_, err := repo.FindByID(ctx, "nonexistent")
	assert.ErrorIs(t, err, apiDomain.ErrPaymentNotFound)
}

// --- FindByOrderID ---

func TestPaymentRepository_FindByOrderID(t *testing.T) {
	ctx, repo, order := setupDeps(t)

	p := makePayment("pay-1", order.ID(), "user-1")
	require.NoError(t, repo.Save(ctx, p))

	found, err := repo.FindByOrderID(ctx, order.ID())
	require.NoError(t, err)
	assert.Equal(t, "pay-1", found.ID())
}

func TestPaymentRepository_FindByOrderID_NotFound(t *testing.T) {
	ctx, repo, _ := setupDeps(t)

	_, err := repo.FindByOrderID(ctx, "nonexistent-order")
	assert.ErrorIs(t, err, apiDomain.ErrPaymentNotFound)
}

// --- UpdateFromMP ---

func TestPaymentRepository_UpdateFromMP(t *testing.T) {
	ctx, repo, order := setupDeps(t)

	p := makePayment("pay-1", order.ID(), "user-1")
	require.NoError(t, repo.Save(ctx, p))

	p.UpdateFromMP(777, paymentDomain.StatusApproved, "accredited", "debit_card")
	require.NoError(t, repo.UpdateFromMP(ctx, p))

	found, err := repo.FindByID(ctx, "pay-1")
	require.NoError(t, err)
	assert.Equal(t, paymentDomain.StatusApproved, found.Status())
	assert.Equal(t, "accredited", found.StatusDetail())
	assert.Equal(t, "debit_card", found.PaymentMethod())
	assert.Equal(t, int64(777), found.MPPaymentID())
}

func TestPaymentRepository_UpdateFromMP_NotFound(t *testing.T) {
	ctx, repo, _ := setupDeps(t)

	// Build a payment that was never saved to the DB.
	p := makePayment("ghost-pay", "order-1", "user-1")
	p.UpdateFromMP(1, paymentDomain.StatusApproved, "accredited", "credit_card")

	err := repo.UpdateFromMP(ctx, p)
	assert.ErrorIs(t, err, apiDomain.ErrPaymentNotFound)
}
