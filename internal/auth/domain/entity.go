package domain

import "errors"

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("password must be at least 8 characters")
	ErrInvalidAccountID   = errors.New("account ID cannot be empty")
	ErrInvalidAccountType = errors.New("invalid account type")
	ErrAuthNotFound       = errors.New("auth not found")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
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

func NewAuth(id string, email, password string, accountID string, accountType AccountType) (*Auth, error) {
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}
	if len(password) < 8 {
		return nil, ErrInvalidPassword
	}
	if accountID == "" {
		return nil, ErrInvalidAccountID
	}
	if !accountType.IsValid() {
		return nil, ErrInvalidAccountType
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
		return ErrInvalidPassword
	}
	a.password = password
	return nil
}

func isValidEmail(email string) bool {
	if len(email) < 3 {
		return false
	}
	hasAt := false
	hasDot := false
	atIndex := -1
	for i, c := range email {
		if c == '@' {
			if hasAt {
				return false
			}
			hasAt = true
			atIndex = i
		}
		if c == '.' && hasAt && i > atIndex+1 {
			hasDot = true
		}
	}
	return hasAt && hasDot
}
