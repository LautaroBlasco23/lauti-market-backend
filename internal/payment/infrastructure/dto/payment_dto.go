package dto

import "time"

type CreatePreferenceRequest struct {
	OrderIDs []string `json:"order_ids" validate:"required,min=1"`
}

type CreateCartPreferenceRequest struct {
	Items []CartItem `json:"items" validate:"required,min=1"`
}

type CartItem struct {
	ProductID string  `json:"product_id" validate:"required"`
	Quantity  int     `json:"quantity" validate:"required,min=1"`
	UnitPrice float64 `json:"unit_price" validate:"required,gt=0"`
}

type CreatePreferenceResponse struct {
	PreferenceID     string `json:"preference_id"`
	InitPoint        string `json:"init_point"`
	SandboxInitPoint string `json:"sandbox_init_point"`
}

type PaymentResponse struct {
	ID            string    `json:"id"`
	OrderID       string    `json:"order_id"`
	UserID        string    `json:"user_id"`
	MPPaymentID   int64     `json:"mp_payment_id,omitempty"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	StatusDetail  string    `json:"status_detail,omitempty"`
	PaymentMethod string    `json:"payment_method,omitempty"`
	PreferenceID  string    `json:"preference_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
