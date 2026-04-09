package domain

import apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"

type OrderItem struct {
	id        string
	orderID   string
	productID string
	quantity  int
	unitPrice float64
}

func NewOrderItem(id, orderID, productID string, quantity int, unitPrice float64) (*OrderItem, error) {
	if quantity <= 0 {
		return nil, apiDomain.ErrInvalidQuantity
	}
	if unitPrice <= 0 {
		return nil, apiDomain.ErrInvalidPrice
	}
	return &OrderItem{
		id:        id,
		orderID:   orderID,
		productID: productID,
		quantity:  quantity,
		unitPrice: unitPrice,
	}, nil
}

func (i *OrderItem) ID() string         { return i.id }
func (i *OrderItem) OrderID() string    { return i.orderID }
func (i *OrderItem) ProductID() string  { return i.productID }
func (i *OrderItem) Quantity() int      { return i.quantity }
func (i *OrderItem) UnitPrice() float64 { return i.unitPrice }
