package domain

import "context"

type MPPreferenceItem struct {
	Title     string
	Quantity  int
	UnitPrice float64
}

type MPBackURLs struct {
	Success string
	Failure string
	Pending string
}

type MPPreferenceRequest struct {
	Items             []MPPreferenceItem
	BackURLs          MPBackURLs
	NotificationURL   string
	ExternalReference string
	MarketplaceFee    float64
	AutoReturn        string
}

type MPPreferenceResponse struct {
	PreferenceID     string
	InitPoint        string
	SandboxInitPoint string
}

type MPPaymentResponse struct {
	ID                int64
	Status            PaymentStatus
	StatusDetail      string
	PaymentMethod     string
	ExternalReference string
}

type MPClient interface {
	CreatePreference(ctx context.Context, req *MPPreferenceRequest) (*MPPreferenceResponse, error)
	CreatePreferenceWithToken(ctx context.Context, accessToken string, req *MPPreferenceRequest) (*MPPreferenceResponse, error)
	GetPayment(ctx context.Context, paymentID int64) (*MPPaymentResponse, error)
}
