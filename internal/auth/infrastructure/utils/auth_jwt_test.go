package utils_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/utils"
)

func TestGenerate_ReturnsValidToken(t *testing.T) {
	g := utils.NewJWTGenerator("secret", time.Hour)
	token, err := g.Generate("auth-1", domain.AccountTypeUser, "acc-1")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerate_ValidateRoundTrip(t *testing.T) {
	g := utils.NewJWTGenerator("secret", time.Hour)
	token, err := g.Generate("auth-1", domain.AccountTypeStore, "acc-42")
	require.NoError(t, err)

	claims, err := g.Validate(token)
	require.NoError(t, err)
	assert.Equal(t, "auth-1", claims.AuthID)
	assert.Equal(t, "acc-42", claims.AccountID)
	assert.Equal(t, domain.AccountTypeStore, claims.AccountType)
}

func TestValidate_ExpiredToken(t *testing.T) {
	g := utils.NewJWTGenerator("secret", time.Hour)

	// Build a token that is already expired.
	claims := utils.Claims{
		AuthID:      "auth-1",
		AccountID:   "acc-1",
		AccountType: domain.AccountTypeUser,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	raw := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := raw.SignedString([]byte("secret"))
	require.NoError(t, err)

	_, err = g.Validate(tokenStr)
	assert.Error(t, err)
}

func TestValidate_InvalidSignature(t *testing.T) {
	g := utils.NewJWTGenerator("secret", time.Hour)
	other := utils.NewJWTGenerator("other-secret", time.Hour)

	token, err := other.Generate("auth-1", domain.AccountTypeUser, "acc-1")
	require.NoError(t, err)

	_, err = g.Validate(token)
	assert.Error(t, err)
}

func TestValidate_MalformedToken(t *testing.T) {
	g := utils.NewJWTGenerator("secret", time.Hour)
	_, err := g.Validate("not.a.valid.jwt")
	assert.Error(t, err)
}
