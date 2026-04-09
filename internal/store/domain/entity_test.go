package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
)

func TestNewStore_Valid(t *testing.T) {
	s, err := domain.NewStore("id-1", "My Store", "Best store", "123 Main St", "555-0100")
	require.NoError(t, err)
	assert.Equal(t, "id-1", s.ID())
	assert.Equal(t, "My Store", s.Name())
	assert.Equal(t, "Best store", s.Description())
	assert.Equal(t, "123 Main St", s.Address())
	assert.Equal(t, "555-0100", s.PhoneNumber())
}

func TestNewStore_EmptyName(t *testing.T) {
	_, err := domain.NewStore("id-1", "", "Best store", "123 Main St", "555-0100")
	assert.ErrorIs(t, err, domain.ErrInvalidName)
}

func TestNewStore_EmptyDescription(t *testing.T) {
	_, err := domain.NewStore("id-1", "My Store", "", "123 Main St", "555-0100")
	assert.ErrorIs(t, err, domain.ErrInvalidDescription)
}

func TestNewStore_EmptyAddress(t *testing.T) {
	_, err := domain.NewStore("id-1", "My Store", "Best store", "", "555-0100")
	assert.ErrorIs(t, err, domain.ErrInvalidAddress)
}

func TestNewStore_EmptyPhoneNumber(t *testing.T) {
	_, err := domain.NewStore("id-1", "My Store", "Best store", "123 Main St", "")
	assert.ErrorIs(t, err, domain.ErrInvalidPhoneNumber)
}

func TestUpdate_Valid(t *testing.T) {
	s, _ := domain.NewStore("id-1", "My Store", "Best store", "123 Main St", "555-0100")
	err := s.Update("New Name", "New Desc", "456 Oak Ave", "555-9999")
	require.NoError(t, err)
	assert.Equal(t, "New Name", s.Name())
	assert.Equal(t, "New Desc", s.Description())
	assert.Equal(t, "456 Oak Ave", s.Address())
	assert.Equal(t, "555-9999", s.PhoneNumber())
}

func TestUpdate_EmptyName(t *testing.T) {
	s, _ := domain.NewStore("id-1", "My Store", "Best store", "123 Main St", "555-0100")
	err := s.Update("", "New Desc", "456 Oak Ave", "555-9999")
	assert.ErrorIs(t, err, domain.ErrInvalidName)
}
