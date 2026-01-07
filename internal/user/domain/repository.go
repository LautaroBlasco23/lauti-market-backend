package domain

import "context"

type Repository interface {
	Save(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id ID) (*User, error)
	Delete(ctx context.Context, id ID) error
}
