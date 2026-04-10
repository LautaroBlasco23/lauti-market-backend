package domain

import (
	"errors"
	"time"
)

// MPFields holds MercadoPago fields for hydration from database
type MPFields struct {
	MPUserID         string
	MPAccessToken    string
	MPRefreshToken   string
	MPTokenExpiresAt *time.Time
	MPConnectedAt    *time.Time
}

// HydrateStore restores MP fields after loading from database
func HydrateStore(store *Store, fields MPFields) *Store {
	store.mpUserID = fields.MPUserID
	store.mpAccessToken = fields.MPAccessToken
	store.mpRefreshToken = fields.MPRefreshToken
	store.mpTokenExpiresAt = fields.MPTokenExpiresAt
	store.mpConnectedAt = fields.MPConnectedAt
	return store
}

var (
	ErrInvalidName        = errors.New("name cannot be empty")
	ErrInvalidDescription = errors.New("description cannot be empty")
	ErrInvalidAddress     = errors.New("address cannot be empty")
	ErrInvalidPhoneNumber = errors.New("phone number cannot be empty")
	ErrStoreNotFound      = errors.New("store not found")
	ErrMPNotConnected     = errors.New("mercadopago account not connected")
	ErrMPInvalidToken     = errors.New("invalid mercadopago token")
)

type Store struct {
	id               string
	name             string
	description      string
	address          string
	phoneNumber      string
	mpUserID         string
	mpAccessToken    string
	mpRefreshToken   string
	mpTokenExpiresAt *time.Time
	mpConnectedAt    *time.Time
}

type CreateStoreInput struct {
	Name        string
	Description string
	Address     string
	PhoneNumber string
}

func NewStore(id string, input CreateStoreInput) (*Store, error) {
	if input.Name == "" {
		return nil, ErrInvalidName
	}
	if input.Description == "" {
		return nil, ErrInvalidDescription
	}
	if input.Address == "" {
		return nil, ErrInvalidAddress
	}
	if input.PhoneNumber == "" {
		return nil, ErrInvalidPhoneNumber
	}
	return &Store{
		id:          id,
		name:        input.Name,
		description: input.Description,
		address:     input.Address,
		phoneNumber: input.PhoneNumber,
	}, nil
}

func (s *Store) ID() string          { return s.id }
func (s *Store) Name() string        { return s.name }
func (s *Store) Description() string { return s.description }
func (s *Store) Address() string     { return s.address }
func (s *Store) PhoneNumber() string { return s.phoneNumber }

// MercadoPago getters
func (s *Store) MPUserID() string             { return s.mpUserID }
func (s *Store) MPAccessToken() string        { return s.mpAccessToken }
func (s *Store) MPRefreshToken() string       { return s.mpRefreshToken }
func (s *Store) MPTokenExpiresAt() *time.Time { return s.mpTokenExpiresAt }
func (s *Store) MPConnectedAt() *time.Time    { return s.mpConnectedAt }

// IsMPConnected returns true if the store has a valid MercadoPago connection
func (s *Store) IsMPConnected() bool {
	return s.mpUserID != "" && s.mpAccessToken != "" && s.mpConnectedAt != nil
}

// IsMPTokenValid returns true if the access token is not expired
func (s *Store) IsMPTokenValid() bool {
	if s.mpTokenExpiresAt == nil {
		return false
	}
	return time.Now().Before(*s.mpTokenExpiresAt)
}

// ConnectMP links the store to a MercadoPago account
func (s *Store) ConnectMP(userID, accessToken, refreshToken string, expiresAt time.Time) {
	s.mpUserID = userID
	s.mpAccessToken = accessToken
	s.mpRefreshToken = refreshToken
	s.mpTokenExpiresAt = &expiresAt
	now := time.Now()
	s.mpConnectedAt = &now
}

// UpdateMPTokens updates the OAuth tokens (used after refresh)
func (s *Store) UpdateMPTokens(accessToken, refreshToken string, expiresAt time.Time) {
	s.mpAccessToken = accessToken
	s.mpRefreshToken = refreshToken
	s.mpTokenExpiresAt = &expiresAt
}

// DisconnectMP removes the MercadoPago connection
func (s *Store) DisconnectMP() {
	s.mpUserID = ""
	s.mpAccessToken = ""
	s.mpRefreshToken = ""
	s.mpTokenExpiresAt = nil
	s.mpConnectedAt = nil
}

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
