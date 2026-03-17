package mp

import (
	"context"
	"fmt"

	"github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/payment"

	domain "github.com/LautaroBlasco23/lauti-market-backend/internal/payment/domain"
)

type mpClient struct {
	client payment.Client
}

func NewMPClient(accessToken string) (domain.MPClient, error) {
	cfg, err := config.New(accessToken)
	if err != nil {
		return nil, fmt.Errorf("creating MP config: %w", err)
	}
	return &mpClient{client: payment.NewClient(cfg)}, nil
}

func (c *mpClient) CreatePayment(ctx context.Context, req *domain.MPPaymentRequest) (*domain.MPPaymentResponse, error) {
	installments := req.Installments
	if installments == 0 {
		installments = 1
	}

	mpReq := payment.Request{
		TransactionAmount: req.Amount,
		Token:             req.CardToken,
		Description:       req.Description,
		Installments:      installments,
		Payer: &payment.PayerRequest{
			Email:     req.PayerEmail,
			FirstName: req.PayerFirstName,
			LastName:  req.PayerLastName,
			Identification: &payment.IdentificationRequest{
				Type:   "DNI",
				Number: req.PayerDNI,
			},
		},
	}

	resp, err := c.client.Create(ctx, mpReq)
	if err != nil {
		return nil, err
	}

	return mapResponse(resp), nil
}

func (c *mpClient) GetPayment(ctx context.Context, paymentID int64) (*domain.MPPaymentResponse, error) {
	resp, err := c.client.Get(ctx, int(paymentID))
	if err != nil {
		return nil, err
	}
	return mapResponse(resp), nil
}

func mapResponse(resp *payment.Response) *domain.MPPaymentResponse {
	return &domain.MPPaymentResponse{
		ID:            int64(resp.ID),
		Status:        mapStatus(resp.Status),
		StatusDetail:  resp.StatusDetail,
		PaymentMethod: resp.PaymentMethodID,
	}
}

func mapStatus(status string) domain.PaymentStatus {
	switch status {
	case "approved":
		return domain.StatusApproved
	case "rejected":
		return domain.StatusRejected
	case "cancelled":
		return domain.StatusCancelled
	case "in_process":
		return domain.StatusInProcess
	default:
		return domain.StatusPending
	}
}
