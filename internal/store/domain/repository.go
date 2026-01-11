package domain

import "context"

type Repository interface {
	Save(ctx context.Context, store *Store) error
	FindByID(ctx context.Context, id string) (*Store, error)
	FindAll(ctx context.Context, limit, offset int) ([]*Store, error)
	Update(ctx context.Context, store *Store) error
	Delete(ctx context.Context, id string) error
}
