//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	orderDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/order/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/order/infrastructure/repository"
	productDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/product/domain"
	productRepo "github.com/LautaroBlasco23/lauti-market-backend/internal/product/infrastructure/repository"
	storeDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
	storeRepo "github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure/repository"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/testutil"
	userDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/user/domain"
	userRepo "github.com/LautaroBlasco23/lauti-market-backend/internal/user/infrastructure/repository"
)

func makeTestOrder(orderID, userID, storeID, productID string) *orderDomain.Order {
	item, _ := orderDomain.NewOrderItem(orderID+"-item-1", orderID, productID, 2, 50.0)
	order, _ := orderDomain.NewOrder(orderID, userID, storeID, []*orderDomain.OrderItem{item}, 100.0)
	return order
}

func TestOrderRepository_Save_FindByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	ctx := context.Background()

	// Seed FK dependencies
	ur := userRepo.NewUserPostgresRepository(db)
	user, _ := userDomain.NewUser("user-1", "John", "Doe")
	require.NoError(t, ur.Save(ctx, user))

	sr := storeRepo.NewStorePostgresRepository(db)
	store, _ := storeDomain.NewStore("store-1", "Test Store", "A test store description", "123 St", "12345678")
	require.NoError(t, sr.Save(ctx, store))

	pr := productRepo.NewProductPostgresRepository(db)
	product, _ := productDomain.NewProduct("prod-1", "store-1", "Widget", "A test product description here", "electronics", 10, 50.0, nil)
	require.NoError(t, pr.Save(ctx, product))

	or := repository.NewOrderPostgresRepository(db)
	order := makeTestOrder("order-1", "user-1", "store-1", "prod-1")
	require.NoError(t, or.Save(ctx, order))

	found, err := or.FindByID(ctx, "order-1")
	require.NoError(t, err)
	assert.Equal(t, "order-1", found.ID())
	assert.Equal(t, "user-1", found.UserID())
	assert.Equal(t, "store-1", found.StoreID())
	assert.Equal(t, orderDomain.StatusPending, found.Status())
	assert.Len(t, found.Items(), 1)

	// Stock should have been decremented by 2 (quantity)
	updatedProduct, err := pr.FindByID(ctx, "prod-1")
	require.NoError(t, err)
	assert.Equal(t, 8, updatedProduct.Stock())
}

func TestOrderRepository_FindByID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	or := repository.NewOrderPostgresRepository(db)
	ctx := context.Background()

	_, err := or.FindByID(ctx, "nonexistent")
	assert.ErrorIs(t, err, apiDomain.ErrOrderNotFound)
}

func TestOrderRepository_FindByUserID_Pagination(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	ctx := context.Background()

	ur := userRepo.NewUserPostgresRepository(db)
	user, _ := userDomain.NewUser("user-1", "John", "Doe")
	require.NoError(t, ur.Save(ctx, user))

	sr := storeRepo.NewStorePostgresRepository(db)
	store, _ := storeDomain.NewStore("store-1", "Test Store", "A test store description", "123 St", "12345678")
	require.NoError(t, sr.Save(ctx, store))

	pr := productRepo.NewProductPostgresRepository(db)
	product, _ := productDomain.NewProduct("prod-1", "store-1", "Widget", "A test product description here", "electronics", 100, 50.0, nil)
	require.NoError(t, pr.Save(ctx, product))

	or := repository.NewOrderPostgresRepository(db)

	for i := 1; i <= 3; i++ {
		id := "order-" + string(rune('0'+i))
		order := makeTestOrder(id, "user-1", "store-1", "prod-1")
		require.NoError(t, or.Save(ctx, order))
	}

	page1, err := or.FindByUserID(ctx, "user-1", 2, 0)
	require.NoError(t, err)
	assert.Len(t, page1, 2)

	page2, err := or.FindByUserID(ctx, "user-1", 2, 2)
	require.NoError(t, err)
	assert.Len(t, page2, 1)
}

func TestOrderRepository_FindByStoreID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	ctx := context.Background()

	ur := userRepo.NewUserPostgresRepository(db)
	user, _ := userDomain.NewUser("user-1", "John", "Doe")
	require.NoError(t, ur.Save(ctx, user))

	sr := storeRepo.NewStorePostgresRepository(db)
	store, _ := storeDomain.NewStore("store-1", "Test Store", "A test store description", "123 St", "12345678")
	require.NoError(t, sr.Save(ctx, store))

	pr := productRepo.NewProductPostgresRepository(db)
	product, _ := productDomain.NewProduct("prod-1", "store-1", "Widget", "A test product description here", "electronics", 50, 50.0, nil)
	require.NoError(t, pr.Save(ctx, product))

	or := repository.NewOrderPostgresRepository(db)
	order := makeTestOrder("order-1", "user-1", "store-1", "prod-1")
	require.NoError(t, or.Save(ctx, order))

	orders, err := or.FindByStoreID(ctx, "store-1", 10, 0)
	require.NoError(t, err)
	assert.Len(t, orders, 1)
	assert.Equal(t, "store-1", orders[0].StoreID())
}

func TestOrderRepository_UpdateStatus(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	ctx := context.Background()

	ur := userRepo.NewUserPostgresRepository(db)
	user, _ := userDomain.NewUser("user-1", "John", "Doe")
	require.NoError(t, ur.Save(ctx, user))

	sr := storeRepo.NewStorePostgresRepository(db)
	store, _ := storeDomain.NewStore("store-1", "Test Store", "A test store description", "123 St", "12345678")
	require.NoError(t, sr.Save(ctx, store))

	pr := productRepo.NewProductPostgresRepository(db)
	product, _ := productDomain.NewProduct("prod-1", "store-1", "Widget", "A test product description here", "electronics", 10, 50.0, nil)
	require.NoError(t, pr.Save(ctx, product))

	or := repository.NewOrderPostgresRepository(db)
	order := makeTestOrder("order-1", "user-1", "store-1", "prod-1")
	require.NoError(t, or.Save(ctx, order))

	require.NoError(t, or.UpdateStatus(ctx, "order-1", orderDomain.StatusConfirmed))

	found, err := or.FindByID(ctx, "order-1")
	require.NoError(t, err)
	assert.Equal(t, orderDomain.StatusConfirmed, found.Status())
}

func TestOrderRepository_UpdateStatus_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	or := repository.NewOrderPostgresRepository(db)
	ctx := context.Background()

	err := or.UpdateStatus(ctx, "nonexistent", orderDomain.StatusConfirmed)
	assert.ErrorIs(t, err, apiDomain.ErrOrderNotFound)
}
