//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/testutil"
	userDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/user/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/infrastructure/repository"
)

func TestUserRepository_Save_FindByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	repo := repository.NewUserPostgresRepository(db)
	ctx := context.Background()

	user, err := userDomain.NewUser("user-1", "John", "Doe")
	require.NoError(t, err)

	require.NoError(t, repo.Save(ctx, user))

	found, err := repo.FindByID(ctx, "user-1")
	require.NoError(t, err)
	assert.Equal(t, "John", found.FirstName())
	assert.Equal(t, "Doe", found.LastName())
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	repo := repository.NewUserPostgresRepository(db)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "nonexistent")
	assert.ErrorIs(t, err, userDomain.ErrUserNotFound)
}

func TestUserRepository_Update(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	repo := repository.NewUserPostgresRepository(db)
	ctx := context.Background()

	user, _ := userDomain.NewUser("user-1", "John", "Doe")
	require.NoError(t, repo.Save(ctx, user))

	require.NoError(t, user.UpdateName("Jane", "Smith"))
	require.NoError(t, repo.Update(ctx, user))

	found, err := repo.FindByID(ctx, "user-1")
	require.NoError(t, err)
	assert.Equal(t, "Jane", found.FirstName())
	assert.Equal(t, "Smith", found.LastName())
}

func TestUserRepository_Update_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	repo := repository.NewUserPostgresRepository(db)
	ctx := context.Background()

	ghost, _ := userDomain.NewUser("ghost-user", "Ghost", "User")
	err := repo.Update(ctx, ghost)
	assert.ErrorIs(t, err, userDomain.ErrUserNotFound)
}

func TestUserRepository_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	repo := repository.NewUserPostgresRepository(db)
	ctx := context.Background()

	user, _ := userDomain.NewUser("user-1", "John", "Doe")
	require.NoError(t, repo.Save(ctx, user))

	require.NoError(t, repo.Delete(ctx, "user-1"))

	_, err := repo.FindByID(ctx, "user-1")
	assert.ErrorIs(t, err, userDomain.ErrUserNotFound)
}

func TestUserRepository_Delete_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	repo := repository.NewUserPostgresRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "nonexistent")
	assert.ErrorIs(t, err, userDomain.ErrUserNotFound)
}
