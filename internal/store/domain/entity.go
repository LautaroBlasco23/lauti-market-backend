package domain

import "errors"

var (
	ErrInvalidName        = errors.New("name cannot be empty")
	ErrInvalidDescription = errors.New("description cannot be empty")
	ErrInvalidAddress     = errors.New("address cannot be empty")
	ErrInvalidPhoneNumber = errors.New("phone number cannot be empty")
	ErrStoreNotFound      = errors.New("store not found")
)

type ID string

type Store struct {
	id          ID
	name        string
	description string
	address     string
	phoneNumber string
}

func NewStore(id ID, name, description, address, phoneNumber string) (*Store, error) {
	if name == "" {
		return nil, ErrInvalidName
	}
	if description == "" {
		return nil, ErrInvalidDescription
	}
	if address == "" {
		return nil, ErrInvalidAddress
	}
	if phoneNumber == "" {
		return nil, ErrInvalidPhoneNumber
	}
	return &Store{
		id:          id,
		name:        name,
		description: description,
		address:     address,
		phoneNumber: phoneNumber,
	}, nil
}

func (s *Store) ID() ID              { return s.id }
func (s *Store) Name() string        { return s.name }
func (s *Store) Description() string { return s.description }
func (s *Store) Address() string     { return s.address }
func (s *Store) PhoneNumber() string { return s.phoneNumber }

func (s *Store) Update(name, description, address, phoneNumber string) error {
	if name == "" {
		return ErrInvalidName
	}
	if description == "" {
		return ErrInvalidDescription
	}
	if address == "" {
		return ErrInvalidAddress
	}
	if phoneNumber == "" {
		return ErrInvalidPhoneNumber
	}
	s.name = name
	s.description = description
	s.address = address
	s.phoneNumber = phoneNumber
	return nil
}
