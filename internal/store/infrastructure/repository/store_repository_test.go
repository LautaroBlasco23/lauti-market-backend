//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	storeDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure/repository"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/testutil"
)

func newStore(id string) *storeDomain.Store {
	s, _ := storeDomain.NewStore(id, "Test Store "+id, "A test store description text", "123 Test St", "123456789")
	return s
}

func TestStoreRepository_Save_FindByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	repo := repository.NewStorePostgresRepository(db)
	ctx := context.Background()

	store := newStore("store-1")
	require.NoError(t, repo.Save(ctx, store))

	found, err := repo.FindByID(ctx, "store-1")
	require.NoError(t, err)
	assert.Equal(t, "store-1", found.ID())
	assert.Equal(t, "Test Store store-1", found.Name())
}

func TestStoreRepository_FindByID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	repo := repository.NewStorePostgresRepository(db)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "nonexistent")
	assert.ErrorIs(t, err, storeDomain.ErrStoreNotFound)
}

func TestStoreRepository_FindAll_Pagination(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	repo := repository.NewStorePostgresRepository(db)
	ctx := context.Background()

	for i := 1; i <= 5; i++ {
		s, _ := storeDomain.NewStore(
			"store-"+string(rune('0'+i)),
			"Store Name "+string(rune('A'+i-1)),
			"A test store description text",
			"123 Test St",
			"123456789",
		)
		require.NoError(t, repo.Save(ctx, s))
	}

	// Page 1: limit 2, offset 0
	page1, err := repo.FindAll(ctx, 2, 0)
	require.NoError(t, err)
	assert.Len(t, page1, 2)

	// Page 2: limit 2, offset 2
	page2, err := repo.FindAll(ctx, 2, 2)
	require.NoError(t, err)
	assert.Len(t, page2, 2)
}

func TestStoreRepository_Update(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	repo := repository.NewStorePostgresRepository(db)
	ctx := context.Background()

	store := newStore("store-1")
	require.NoError(t, repo.Save(ctx, store))

	require.NoError(t, store.Update("Updated Name", "Updated description text here", "456 New St", "987654321"))
	require.NoError(t, repo.Update(ctx, store))

	found, err := repo.FindByID(ctx, "store-1")
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", found.Name())
	assert.Equal(t, "456 New St", found.Address())
}

func TestStoreRepository_Update_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	repo := repository.NewStorePostgresRepository(db)
	ctx := context.Background()

	ghost := newStore("ghost-store")
	err := repo.Update(ctx, ghost)
	assert.ErrorIs(t, err, storeDomain.ErrStoreNotFound)
}

func TestStoreRepository_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	repo := repository.NewStorePostgresRepository(db)
	ctx := context.Background()

	store := newStore("store-1")
	require.NoError(t, repo.Save(ctx, store))

	require.NoError(t, repo.Delete(ctx, "store-1"))

	_, err := repo.FindByID(ctx, "store-1")
	assert.ErrorIs(t, err, storeDomain.ErrStoreNotFound)
}
