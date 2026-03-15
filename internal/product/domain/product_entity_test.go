package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/product/domain"
)

func TestNewProduct_Valid(t *testing.T) {
	p, err := domain.NewProduct("id-1", "store-1", "Widget", "A widget", "tools", 10, 9.99, nil)
	require.NoError(t, err)
	assert.Equal(t, "id-1", p.ID())
	assert.Equal(t, "store-1", p.StoreID())
	assert.Equal(t, "Widget", p.Name())
	assert.Nil(t, p.ImageURL())
}

func TestNewProduct_WithImageURL(t *testing.T) {
	url := "https://example.com/img.jpg"
	p, err := domain.NewProduct("id-1", "store-1", "Widget", "A widget", "tools", 10, 9.99, &url)
	require.NoError(t, err)
	require.NotNil(t, p.ImageURL())
	assert.Equal(t, url, *p.ImageURL())
}

func TestNewProduct_EmptyStoreID(t *testing.T) {
	_, err := domain.NewProduct("id-1", "", "Widget", "A widget", "tools", 10, 9.99, nil)
	assert.ErrorIs(t, err, apiDomain.ErrInvalidStoreID)
}

func TestNewProduct_EmptyName(t *testing.T) {
	_, err := domain.NewProduct("id-1", "store-1", "", "A widget", "tools", 10, 9.99, nil)
	assert.ErrorIs(t, err, apiDomain.ErrInvalidProductName)
}

func TestNewProduct_EmptyDescription(t *testing.T) {
	_, err := domain.NewProduct("id-1", "store-1", "Widget", "", "tools", 10, 9.99, nil)
	assert.ErrorIs(t, err, apiDomain.ErrInvalidProductDescription)
}

func TestNewProduct_EmptyCategory(t *testing.T) {
	_, err := domain.NewProduct("id-1", "store-1", "Widget", "A widget", "", 10, 9.99, nil)
	assert.ErrorIs(t, err, apiDomain.ErrInvalidCategory)
}

func TestNewProduct_NegativeStock(t *testing.T) {
	_, err := domain.NewProduct("id-1", "store-1", "Widget", "A widget", "tools", -1, 9.99, nil)
	assert.ErrorIs(t, err, apiDomain.ErrInvalidStock)
}

func TestNewProduct_ZeroPrice(t *testing.T) {
	_, err := domain.NewProduct("id-1", "store-1", "Widget", "A widget", "tools", 10, 0, nil)
	assert.ErrorIs(t, err, apiDomain.ErrInvalidPrice)
}

func TestNewProduct_NegativePrice(t *testing.T) {
	_, err := domain.NewProduct("id-1", "store-1", "Widget", "A widget", "tools", 10, -5.0, nil)
	assert.ErrorIs(t, err, apiDomain.ErrInvalidPrice)
}

func TestUpdate_Valid(t *testing.T) {
	p, _ := domain.NewProduct("id-1", "store-1", "Widget", "A widget", "tools", 10, 9.99, nil)
	url := "https://example.com/img.jpg"
	err := p.Update("Gadget", "A gadget", "electronics", 5, 19.99, &url)
	require.NoError(t, err)
	assert.Equal(t, "Gadget", p.Name())
	assert.Equal(t, "electronics", p.Category())
	assert.Equal(t, 5, p.Stock())
	assert.Equal(t, 19.99, p.Price())
}

func TestUpdate_NegativeStock(t *testing.T) {
	p, _ := domain.NewProduct("id-1", "store-1", "Widget", "A widget", "tools", 10, 9.99, nil)
	err := p.Update("Widget", "A widget", "tools", -1, 9.99, nil)
	assert.ErrorIs(t, err, apiDomain.ErrInvalidStock)
}
