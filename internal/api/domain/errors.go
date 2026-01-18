package domain

import "errors"

type InvalidPayloadError struct {
	Fields map[string]string
}

func (e InvalidPayloadError) Error() string {
	return "invalid payload"
}

var (
	ErrInvalidProductName        = errors.New("product name cannot be empty")
	ErrInvalidProductDescription = errors.New("product description cannot be empty")
	ErrInvalidStock              = errors.New("stock cannot be negative")
	ErrInvalidPrice              = errors.New("price must be greater than zero")
	ErrInvalidStoreID            = errors.New("store id cannot be empty")
	ErrProductNotFound           = errors.New("product not found")
)
