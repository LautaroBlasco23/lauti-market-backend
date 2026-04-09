//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	authDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/auth/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/repository"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/testutil"
)

func TestAuthRepository_Save_FindByEmail(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	repo := repository.NewPostgresRepository(db)
	ctx := context.Background()

	auth, err := authDomain.NewAuth("auth-1", "test@example.com", "password123", "acc-1", authDomain.AccountTypeUser)
	require.NoError(t, err)

	require.NoError(t, repo.Save(ctx, auth))

	found, err := repo.FindByEmail(ctx, "test@example.com")
	require.NoError(t, err)
	assert.Equal(t, "auth-1", found.ID())
	assert.Equal(t, "test@example.com", found.Email())
	assert.Equal(t, authDomain.AccountTypeUser, found.AccountType())
}

func TestAuthRepository_Save_DuplicateEmail(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	repo := repository.NewPostgresRepository(db)
	ctx := context.Background()

	auth1, _ := authDomain.NewAuth("auth-1", "dup@example.com", "password123", "acc-1", authDomain.AccountTypeUser)
	auth2, _ := authDomain.NewAuth("auth-2", "dup@example.com", "password123", "acc-2", authDomain.AccountTypeUser)

	require.NoError(t, repo.Save(ctx, auth1))
	err := repo.Save(ctx, auth2)
	assert.ErrorIs(t, err, apiDomain.ErrEmailExists)
}

func TestAuthRepository_FindByEmail_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	repo := repository.NewPostgresRepository(db)
	ctx := context.Background()

	_, err := repo.FindByEmail(ctx, "nobody@example.com")
	assert.ErrorIs(t, err, apiDomain.ErrAuthNotFound)
}

func TestAuthRepository_FindByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	repo := repository.NewPostgresRepository(db)
	ctx := context.Background()

	auth, _ := authDomain.NewAuth("auth-1", "id@example.com", "password123", "acc-1", authDomain.AccountTypeStore)
	require.NoError(t, repo.Save(ctx, auth))

	found, err := repo.FindByID(ctx, "auth-1")
	require.NoError(t, err)
	assert.Equal(t, authDomain.AccountTypeStore, found.AccountType())
}

func TestAuthRepository_FindByID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	repo := repository.NewPostgresRepository(db)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "nonexistent")
	assert.ErrorIs(t, err, apiDomain.ErrAuthNotFound)
}

func TestAuthRepository_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	repo := repository.NewPostgresRepository(db)
	ctx := context.Background()

	auth, _ := authDomain.NewAuth("auth-1", "del@example.com", "password123", "acc-1", authDomain.AccountTypeUser)
	require.NoError(t, repo.Save(ctx, auth))

	require.NoError(t, repo.Delete(ctx, "auth-1"))

	_, err := repo.FindByID(ctx, "auth-1")
	assert.ErrorIs(t, err, apiDomain.ErrAuthNotFound)
}

func TestAuthRepository_Delete_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)
	repo := repository.NewPostgresRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "nonexistent")
	assert.ErrorIs(t, err, apiDomain.ErrAuthNotFound)
}
