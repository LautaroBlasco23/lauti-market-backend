package dto

import "time"

type OrderItemRequest struct {
	ProductID string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity"   validate:"required,min=1"`
}

type CreateOrderRequest struct {
	StoreID string             `json:"store_id" validate:"required"`
	Items   []OrderItemRequest `json:"items"    validate:"required,min=1,dive"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

type OrderItemResponse struct {
	ID        string  `json:"id"`
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type OrderResponse struct {
	ID         string              `json:"id"`
	UserID     string              `json:"user_id"`
	StoreID    string              `json:"store_id"`
	Status     string              `json:"status"`
	Items      []OrderItemResponse `json:"items"`
	TotalPrice float64             `json:"total_price"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
}
