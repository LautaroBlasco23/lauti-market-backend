package domain

import "errors"

type InvalidPayloadError struct {
	Fields map[string]string
}

func (e InvalidPayloadError) Error() string {
	return "invalid payload"
}

var (
	// Auth errors
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("password must be at least 8 characters")
	ErrInvalidAccountID   = errors.New("account ID cannot be empty")
	ErrInvalidAccountType = errors.New("invalid account type")
	ErrAuthNotFound       = errors.New("auth not found")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")

	// Product errors
	ErrInvalidProductName        = errors.New("product name cannot be empty")
	ErrInvalidProductDescription = errors.New("product description cannot be empty")
	ErrInvalidStock              = errors.New("stock cannot be negative")
	ErrInvalidPrice              = errors.New("price must be greater than zero")
	ErrInvalidStoreID            = errors.New("store id cannot be empty")
	ErrInvalidCategory           = errors.New("category cannot be empty")
	ErrProductNotFound           = errors.New("product not found")

	// Order errors
	ErrOrderNotFound            = errors.New("order not found")
	ErrInsufficientStock        = errors.New("insufficient stock")
	ErrEmptyOrderItems          = errors.New("order must have at least one item")
	ErrInvalidQuantity          = errors.New("quantity must be greater than zero")
	ErrInvalidOrderStatus       = errors.New("invalid order status")
	ErrForbiddenTransition      = errors.New("invalid status transition")
	ErrUnauthorized             = errors.New("unauthorized")
	ErrForbidden                = errors.New("forbidden")
	ErrItemsFromMultipleStores  = errors.New("all items must belong to the same store")
)
