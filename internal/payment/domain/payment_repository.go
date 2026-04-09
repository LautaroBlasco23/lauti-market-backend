package domain

import "context"

type Repository interface {
	Save(ctx context.Context, payment *Payment) error
	FindByID(ctx context.Context, id string) (*Payment, error)
	FindByOrderID(ctx context.Context, orderID string) (*Payment, error)
	UpdateFromMP(ctx context.Context, payment *Payment) error
}
