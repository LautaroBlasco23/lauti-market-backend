package domain

import "context"

type Repository interface {
	Save(ctx context.Context, auth *Auth) error
	FindByID(ctx context.Context, id string) (*Auth, error)
	FindByEmail(ctx context.Context, email string) (*Auth, error)
	Delete(ctx context.Context, id string) error
}
