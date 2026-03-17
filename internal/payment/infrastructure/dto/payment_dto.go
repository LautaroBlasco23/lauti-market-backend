package dto

import "time"

type CreatePaymentRequest struct {
	OrderID      string `json:"order_id"     validate:"required"`
	CardToken    string `json:"card_token"   validate:"required"`
	PayerEmail   string `json:"payer_email"  validate:"required,email"`
	Installments int    `json:"installments" validate:"omitempty,min=1,max=12"`
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
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
