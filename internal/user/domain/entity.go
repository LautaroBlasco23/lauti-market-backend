package domain

import "errors"

var (
	ErrInvalidFirstName = errors.New("first name cannot be empty")
	ErrInvalidLastName  = errors.New("last name cannot be empty")
	ErrUserNotFound     = errors.New("user not found")
)

type ID string

type User struct {
	id        ID
	firstName string
	lastName  string
}

func NewUser(id ID, firstName, lastName string) (*User, error) {
	if firstName == "" {
		return nil, ErrInvalidFirstName
	}
	if lastName == "" {
		return nil, ErrInvalidLastName
	}
	return &User{
		id:        id,
		firstName: firstName,
		lastName:  lastName,
	}, nil
}

func (u *User) ID() ID            { return u.id }
func (u *User) FirstName() string { return u.firstName }
func (u *User) LastName() string  { return u.lastName }

func (u *User) UpdateName(firstName, lastName string) error {
	if firstName == "" {
		return ErrInvalidFirstName
	}
	if lastName == "" {
		return ErrInvalidLastName
	}
	u.firstName = firstName
	u.lastName = lastName
	return nil
}
