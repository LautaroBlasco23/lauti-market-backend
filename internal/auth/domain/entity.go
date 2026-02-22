package domain

import (
	"net/mail"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
)

type AccountType string

const (
	AccountTypeUser  AccountType = "user"
	AccountTypeStore AccountType = "store"
)

func (at AccountType) IsValid() bool {
	return at == AccountTypeUser || at == AccountTypeStore
}

type Auth struct {
	id          string
	email       string
	password    string
	accountID   string
	accountType AccountType
}

func NewAuth(id, email, password, accountID string, accountType AccountType) (*Auth, error) {
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, apiDomain.ErrInvalidEmail
	}
	if len(password) < 8 {
		return nil, apiDomain.ErrInvalidPassword
	}
	if accountID == "" {
		return nil, apiDomain.ErrInvalidAccountID
	}
	if !accountType.IsValid() {
		return nil, apiDomain.ErrInvalidAccountType
	}
	return &Auth{
		id:          id,
		email:       email,
		password:    password,
		accountID:   accountID,
		accountType: accountType,
	}, nil
}

func (a *Auth) ID() string               { return a.id }
func (a *Auth) Email() string            { return a.email }
func (a *Auth) Password() string         { return a.password }
func (a *Auth) AccountID() string        { return a.accountID }
func (a *Auth) AccountType() AccountType { return a.accountType }

func (a *Auth) UpdatePassword(password string) error {
	if len(password) < 8 {
		return apiDomain.ErrInvalidPassword
	}
	a.password = password
	return nil
}
