package domain

import "errors"

var (
	// User Input related
	ErrInvalidFirstName = errors.New("first name must be between 2 and 50 characters")
	ErrInvalidLastName  = errors.New("last name must be between 2 and 50 characters")

	// Store Input related
	ErrInvalidStoreName        = errors.New("store name must be between 2 and 100 characters")
	ErrInvalidStoreDescription = errors.New("store description must be between 10 and 500 characters")
	ErrInvalidStoreAddress     = errors.New("store address must be between 5 and 200 characters")
	ErrInvalidPhoneNumber      = errors.New("phone number must be between 8 and 20 characters")
)
