package domain

import "errors"

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("password must be at least 8 characters")
	ErrInvalidUserID      = errors.New("user ID cannot be empty")
	ErrAuthNotFound       = errors.New("auth not found")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type ID string
type UserID string

type Auth struct {
	id       ID
	email    string
	password string
	userID   UserID
}

func NewAuth(id ID, email, password string, userID UserID) (*Auth, error) {
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}
	if len(password) < 8 {
		return nil, ErrInvalidPassword
	}
	if userID == "" {
		return nil, ErrInvalidUserID
	}
	return &Auth{
		id:       id,
		email:    email,
		password: password,
		userID:   userID,
	}, nil
}

func (a *Auth) ID() ID           { return a.id }
func (a *Auth) Email() string    { return a.email }
func (a *Auth) Password() string { return a.password }
func (a *Auth) UserID() UserID   { return a.userID }

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
