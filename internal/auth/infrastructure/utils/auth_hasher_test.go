package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/utils"
)

func TestHash_ReturnsHashedPassword(t *testing.T) {
	h := utils.NewBcryptHasher()
	hash, err := h.Hash("mypassword")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "mypassword", hash)
}

func TestHash_DifferentHashesForSamePassword(t *testing.T) {
	h := utils.NewBcryptHasher()
	hash1, err := h.Hash("mypassword")
	require.NoError(t, err)
	hash2, err := h.Hash("mypassword")
	require.NoError(t, err)
	assert.NotEqual(t, hash1, hash2)
}

func TestCompare_CorrectPassword(t *testing.T) {
	h := utils.NewBcryptHasher()
	hash, err := h.Hash("mypassword")
	require.NoError(t, err)
	err = h.Compare(hash, "mypassword")
	assert.NoError(t, err)
}

func TestCompare_WrongPassword(t *testing.T) {
	h := utils.NewBcryptHasher()
	hash, err := h.Hash("mypassword")
	require.NoError(t, err)
	err = h.Compare(hash, "wrongpassword")
	assert.Error(t, err)
}
