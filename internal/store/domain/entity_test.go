package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
)

func TestNewStore_Valid(t *testing.T) {
	input := domain.CreateStoreInput{
		Name:        "My Store",
		Description: "Best store",
		Address:     "123 Main St",
		PhoneNumber: "555-0100",
	}
	s, err := domain.NewStore("id-1", input)
	require.NoError(t, err)
	assert.Equal(t, "id-1", s.ID())
	assert.Equal(t, "My Store", s.Name())
	assert.Equal(t, "Best store", s.Description())
	assert.Equal(t, "123 Main St", s.Address())
	assert.Equal(t, "555-0100", s.PhoneNumber())
}

func TestNewStore_EmptyName(t *testing.T) {
	input := domain.CreateStoreInput{
		Name:        "",
		Description: "Best store",
		Address:     "123 Main St",
		PhoneNumber: "555-0100",
	}
	_, err := domain.NewStore("id-1", input)
	assert.ErrorIs(t, err, domain.ErrInvalidName)
}

func TestNewStore_EmptyDescription(t *testing.T) {
	input := domain.CreateStoreInput{
		Name:        "My Store",
		Description: "",
		Address:     "123 Main St",
		PhoneNumber: "555-0100",
	}
	_, err := domain.NewStore("id-1", input)
	assert.ErrorIs(t, err, domain.ErrInvalidDescription)
}

func TestNewStore_EmptyAddress(t *testing.T) {
	input := domain.CreateStoreInput{
		Name:        "My Store",
		Description: "Best store",
		Address:     "",
		PhoneNumber: "555-0100",
	}
	_, err := domain.NewStore("id-1", input)
	assert.ErrorIs(t, err, domain.ErrInvalidAddress)
}

func TestNewStore_EmptyPhoneNumber(t *testing.T) {
	input := domain.CreateStoreInput{
		Name:        "My Store",
		Description: "Best store",
		Address:     "123 Main St",
		PhoneNumber: "",
	}
	_, err := domain.NewStore("id-1", input)
	assert.ErrorIs(t, err, domain.ErrInvalidPhoneNumber)
}

func TestUpdate_Valid(t *testing.T) {
	input := domain.CreateStoreInput{
		Name:        "My Store",
		Description: "Best store",
		Address:     "123 Main St",
		PhoneNumber: "555-0100",
	}
	s, _ := domain.NewStore("id-1", input)
	err := s.Update("New Name", "New Desc", "456 Oak Ave", "555-9999")
	require.NoError(t, err)
	assert.Equal(t, "New Name", s.Name())
	assert.Equal(t, "New Desc", s.Description())
	assert.Equal(t, "456 Oak Ave", s.Address())
	assert.Equal(t, "555-9999", s.PhoneNumber())
}

func TestUpdate_EmptyName(t *testing.T) {
	input := domain.CreateStoreInput{
		Name:        "My Store",
		Description: "Best store",
		Address:     "123 Main St",
		PhoneNumber: "555-0100",
	}
	s, _ := domain.NewStore("id-1", input)
	err := s.Update("", "New Desc", "456 Oak Ave", "555-9999")
	assert.ErrorIs(t, err, domain.ErrInvalidName)
}
