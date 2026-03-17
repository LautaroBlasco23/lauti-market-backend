package domain

import (
	"time"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
)

type Order struct {
	id         string
	userID     string
	storeID    string
	status     OrderStatus
	items      []*OrderItem
	totalPrice float64
	createdAt  time.Time
	updatedAt  time.Time
}

func NewOrder(id, userID, storeID string, items []*OrderItem, totalPrice float64) (*Order, error) {
	if len(items) == 0 {
		return nil, apiDomain.ErrEmptyOrderItems
	}
	if totalPrice <= 0 {
		return nil, apiDomain.ErrInvalidPrice
	}
	return &Order{
		id:         id,
		userID:     userID,
		storeID:    storeID,
		status:     StatusPending,
		items:      items,
		totalPrice: totalPrice,
		createdAt:  time.Now(),
		updatedAt:  time.Now(),
	}, nil
}

// NewOrderFromDB reconstructs an Order from persisted data.
func NewOrderFromDB(id, userID, storeID string, status OrderStatus, items []*OrderItem, totalPrice float64, createdAt, updatedAt time.Time) *Order {
	return &Order{
		id:         id,
		userID:     userID,
		storeID:    storeID,
		status:     status,
		items:      items,
		totalPrice: totalPrice,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}
}

func (o *Order) TransitionStatus(newStatus OrderStatus, accountType, accountID string) error { //nolint:gocyclo
	switch {
	case o.status == StatusPending && newStatus == StatusConfirmed:
		if accountType != "store" || accountID != o.storeID {
			return apiDomain.ErrForbidden
		}
	case o.status == StatusConfirmed && newStatus == StatusShipped:
		if accountType != "store" || accountID != o.storeID {
			return apiDomain.ErrForbidden
		}
	case o.status == StatusShipped && newStatus == StatusDelivered:
		if accountType != "store" || accountID != o.storeID {
			return apiDomain.ErrForbidden
		}
	case o.status == StatusPending && newStatus == StatusCancelled:
		if accountType != "user" || accountID != o.userID {
			return apiDomain.ErrForbidden
		}
	default:
		return apiDomain.ErrForbiddenTransition
	}

	o.status = newStatus
	o.updatedAt = time.Now()
	return nil
}

func (o *Order) ID() string           { return o.id }
func (o *Order) UserID() string       { return o.userID }
func (o *Order) StoreID() string      { return o.storeID }
func (o *Order) Status() OrderStatus  { return o.status }
func (o *Order) Items() []*OrderItem  { return o.items }
func (o *Order) TotalPrice() float64  { return o.totalPrice }
func (o *Order) CreatedAt() time.Time { return o.createdAt }
func (o *Order) UpdatedAt() time.Time { return o.updatedAt }
