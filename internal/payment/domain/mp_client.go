package domain

import "context"

type MPPaymentRequest struct {
	Amount         float64
	Description    string
	PayerEmail     string
	PayerFirstName string
	PayerLastName  string
	PayerDNI       string
	CardToken      string
	Installments   int
	IdempotencyKey string
}

type MPPaymentResponse struct {
	ID            int64
	Status        PaymentStatus
	StatusDetail  string
	PaymentMethod string
}

type MPClient interface {
	CreatePayment(ctx context.Context, req *MPPaymentRequest) (*MPPaymentResponse, error)
	GetPayment(ctx context.Context, paymentID int64) (*MPPaymentResponse, error)
}
