//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	productDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/product/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/product/infrastructure/repository"
	storeDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
	storeRepo "github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure/repository"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/testutil"
)

func TestProductRepository_Save_FindByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	ctx := context.Background()

	// Seed a store (required FK)
	sr := storeRepo.NewStorePostgresRepository(db)
	store, _ := storeDomain.NewStore("store-1", storeDomain.CreateStoreInput{
		Name:        "Test Store",
		Description: "A test store description",
		Address:     "123 St",
		PhoneNumber: "12345678",
	})
	require.NoError(t, sr.Save(ctx, store))

	pr := repository.NewProductPostgresRepository(db)
	product, _ := productDomain.NewProduct("prod-1", "store-1", "Widget", "A test product description here", "electronics", 5, 19.99, nil)
	require.NoError(t, pr.Save(ctx, product))

	found, err := pr.FindByID(ctx, "prod-1")
	require.NoError(t, err)
	assert.Equal(t, "Widget", found.Name())
	assert.Equal(t, "store-1", found.StoreID())
	assert.Equal(t, 5, found.Stock())
}

func TestProductRepository_FindByID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	pr := repository.NewProductPostgresRepository(db)
	ctx := context.Background()

	_, err := pr.FindByID(ctx, "nonexistent")
	assert.ErrorIs(t, err, apiDomain.ErrProductNotFound)
}

func TestProductRepository_FindAll_WithCategory(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	ctx := context.Background()

	sr := storeRepo.NewStorePostgresRepository(db)
	store, _ := storeDomain.NewStore("store-1", storeDomain.CreateStoreInput{
		Name:        "Test Store",
		Description: "A test store description",
		Address:     "123 St",
		PhoneNumber: "12345678",
	})
	require.NoError(t, sr.Save(ctx, store))

	pr := repository.NewProductPostgresRepository(db)
	p1, _ := productDomain.NewProduct("prod-1", "store-1", "Widget A", "A test product description here", "electronics", 5, 9.99, nil)
	p2, _ := productDomain.NewProduct("prod-2", "store-1", "Widget B", "A test product description here", "clothing", 3, 19.99, nil)
	require.NoError(t, pr.Save(ctx, p1))
	require.NoError(t, pr.Save(ctx, p2))

	cat := "electronics"
	products, err := pr.FindAll(ctx, 10, 0, &cat)
	require.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, "electronics", products[0].Category())

	all, err := pr.FindAll(ctx, 10, 0, nil)
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestProductRepository_FindByStoreID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	ctx := context.Background()

	sr := storeRepo.NewStorePostgresRepository(db)
	store1, _ := storeDomain.NewStore("store-1", storeDomain.CreateStoreInput{
		Name:        "Store One",
		Description: "A test store description",
		Address:     "123 St",
		PhoneNumber: "12345678",
	})
	store2, _ := storeDomain.NewStore("store-2", storeDomain.CreateStoreInput{
		Name:        "Store Two",
		Description: "A test store description",
		Address:     "456 St",
		PhoneNumber: "87654321",
	})
	require.NoError(t, sr.Save(ctx, store1))
	require.NoError(t, sr.Save(ctx, store2))

	pr := repository.NewProductPostgresRepository(db)
	p1, _ := productDomain.NewProduct("prod-1", "store-1", "Widget A", "A test product description here", "electronics", 5, 9.99, nil)
	p2, _ := productDomain.NewProduct("prod-2", "store-2", "Widget B", "A test product description here", "clothing", 3, 19.99, nil)
	require.NoError(t, pr.Save(ctx, p1))
	require.NoError(t, pr.Save(ctx, p2))

	products, err := pr.FindByStoreID(ctx, "store-1", 10, 0)
	require.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, "prod-1", products[0].ID())
}

func TestProductRepository_Update(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	ctx := context.Background()

	sr := storeRepo.NewStorePostgresRepository(db)
	store, _ := storeDomain.NewStore("store-1", storeDomain.CreateStoreInput{
		Name:        "Test Store",
		Description: "A test store description",
		Address:     "123 St",
		PhoneNumber: "12345678",
	})
	require.NoError(t, sr.Save(ctx, store))

	pr := repository.NewProductPostgresRepository(db)
	product, _ := productDomain.NewProduct("prod-1", "store-1", "Widget", "A test product description here", "electronics", 5, 9.99, nil)
	require.NoError(t, pr.Save(ctx, product))

	require.NoError(t, product.Update("Updated Widget", "An updated product description here", "electronics", 20, 29.99, nil))
	require.NoError(t, pr.Update(ctx, product))

	found, err := pr.FindByID(ctx, "prod-1")
	require.NoError(t, err)
	assert.Equal(t, "Updated Widget", found.Name())
	assert.Equal(t, 20, found.Stock())
}

func TestProductRepository_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	ctx := context.Background()

	sr := storeRepo.NewStorePostgresRepository(db)
	store, _ := storeDomain.NewStore("store-1", storeDomain.CreateStoreInput{
		Name:        "Test Store",
		Description: "A test store description",
		Address:     "123 St",
		PhoneNumber: "12345678",
	})
	require.NoError(t, sr.Save(ctx, store))

	pr := repository.NewProductPostgresRepository(db)
	product, _ := productDomain.NewProduct("prod-1", "store-1", "Widget", "A test product description here", "electronics", 5, 9.99, nil)
	require.NoError(t, pr.Save(ctx, product))

	require.NoError(t, pr.Delete(ctx, "prod-1"))

	_, err := pr.FindByID(ctx, "prod-1")
	assert.ErrorIs(t, err, apiDomain.ErrProductNotFound)
}
