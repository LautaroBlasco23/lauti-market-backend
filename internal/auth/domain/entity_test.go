package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/domain"
)

func TestNewAuth_ValidInput(t *testing.T) {
	auth, err := domain.NewAuth("id-1", "user@example.com", "password123", "acc-1", domain.AccountTypeUser)
	require.NoError(t, err)
	assert.Equal(t, "id-1", auth.ID())
	assert.Equal(t, "user@example.com", auth.Email())
	assert.Equal(t, "password123", auth.Password())
	assert.Equal(t, "acc-1", auth.AccountID())
	assert.Equal(t, domain.AccountTypeUser, auth.AccountType())
}

func TestNewAuth_InvalidEmail(t *testing.T) {
	_, err := domain.NewAuth("id-1", "not-an-email", "password123", "acc-1", domain.AccountTypeUser)
	assert.ErrorIs(t, err, apiDomain.ErrInvalidEmail)
}

func TestNewAuth_ShortPassword(t *testing.T) {
	_, err := domain.NewAuth("id-1", "user@example.com", "short", "acc-1", domain.AccountTypeUser)
	assert.ErrorIs(t, err, apiDomain.ErrInvalidPassword)
}

func TestNewAuth_EmptyAccountID(t *testing.T) {
	_, err := domain.NewAuth("id-1", "user@example.com", "password123", "", domain.AccountTypeUser)
	assert.ErrorIs(t, err, apiDomain.ErrInvalidAccountID)
}

func TestNewAuth_InvalidAccountType(t *testing.T) {
	_, err := domain.NewAuth("id-1", "user@example.com", "password123", "acc-1", domain.AccountType("admin"))
	assert.ErrorIs(t, err, apiDomain.ErrInvalidAccountType)
}

func TestAccountType_IsValid(t *testing.T) {
	assert.True(t, domain.AccountTypeUser.IsValid())
	assert.True(t, domain.AccountTypeStore.IsValid())
	assert.False(t, domain.AccountType("admin").IsValid())
	assert.False(t, domain.AccountType("").IsValid())
}

func TestUpdatePassword_Valid(t *testing.T) {
	auth, err := domain.NewAuth("id-1", "user@example.com", "password123", "acc-1", domain.AccountTypeUser)
	require.NoError(t, err)
	err = auth.UpdatePassword("newpassword1")
	assert.NoError(t, err)
	assert.Equal(t, "newpassword1", auth.Password())
}

func TestUpdatePassword_TooShort(t *testing.T) {
	auth, err := domain.NewAuth("id-1", "user@example.com", "password123", "acc-1", domain.AccountTypeUser)
	require.NoError(t, err)
	err = auth.UpdatePassword("short")
	assert.ErrorIs(t, err, apiDomain.ErrInvalidPassword)
}
