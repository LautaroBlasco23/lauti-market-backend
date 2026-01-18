package domain

import "context"

type Repository interface {
	Save(ctx context.Context, product *Product) error
	FindByID(ctx context.Context, id string) (*Product, error)
	FindByStoreID(ctx context.Context, storeID string, limit, offset int) ([]*Product, error)
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, id string) error
}
