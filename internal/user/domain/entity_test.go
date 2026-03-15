package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/domain"
)

func TestNewUser_Valid(t *testing.T) {
	u, err := domain.NewUser("id-1", "John", "Doe")
	require.NoError(t, err)
	assert.Equal(t, "id-1", u.ID())
	assert.Equal(t, "John", u.FirstName())
	assert.Equal(t, "Doe", u.LastName())
}

func TestNewUser_EmptyFirstName(t *testing.T) {
	_, err := domain.NewUser("id-1", "", "Doe")
	assert.ErrorIs(t, err, domain.ErrInvalidFirstName)
}

func TestNewUser_EmptyLastName(t *testing.T) {
	_, err := domain.NewUser("id-1", "John", "")
	assert.ErrorIs(t, err, domain.ErrInvalidLastName)
}

func TestUpdateName_Valid(t *testing.T) {
	u, err := domain.NewUser("id-1", "John", "Doe")
	require.NoError(t, err)

	err = u.UpdateName("Jane", "Smith")
	require.NoError(t, err)
	assert.Equal(t, "Jane", u.FirstName())
	assert.Equal(t, "Smith", u.LastName())
}

func TestUpdateName_EmptyFirstName(t *testing.T) {
	u, _ := domain.NewUser("id-1", "John", "Doe")
	err := u.UpdateName("", "Smith")
	assert.ErrorIs(t, err, domain.ErrInvalidFirstName)
}

func TestUpdateName_EmptyLastName(t *testing.T) {
	u, _ := domain.NewUser("id-1", "John", "Doe")
	err := u.UpdateName("Jane", "")
	assert.ErrorIs(t, err, domain.ErrInvalidLastName)
}
