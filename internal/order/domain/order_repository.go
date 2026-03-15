package domain

import "context"

type Repository interface {
	Save(ctx context.Context, order *Order) error
	FindByID(ctx context.Context, id string) (*Order, error)
	FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*Order, error)
	FindByStoreID(ctx context.Context, storeID string, limit, offset int) ([]*Order, error)
	UpdateStatus(ctx context.Context, id string, status OrderStatus) error
}
