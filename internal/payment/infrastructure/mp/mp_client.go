package mp

import (
	"context"
	"fmt"

	"github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/payment"
	"github.com/mercadopago/sdk-go/pkg/preference"

	domain "github.com/LautaroBlasco23/lauti-market-backend/internal/payment/domain"
)

type mpClient struct {
	paymentClient    payment.Client
	preferenceClient preference.Client
}

func NewMPClient(accessToken string) (domain.MPClient, error) {
	cfg, err := config.New(accessToken)
	if err != nil {
		return nil, fmt.Errorf("creating MP config: %w", err)
	}
	return &mpClient{
		paymentClient:    payment.NewClient(cfg),
		preferenceClient: preference.NewClient(cfg),
	}, nil
}

func (c *mpClient) CreatePreference(ctx context.Context, req *domain.MPPreferenceRequest) (*domain.MPPreferenceResponse, error) {
	return c.createPreferenceWithClient(ctx, c.preferenceClient, req)
}

func (c *mpClient) CreatePreferenceWithToken(ctx context.Context, accessToken string, req *domain.MPPreferenceRequest) (*domain.MPPreferenceResponse, error) {
	cfg, err := config.New(accessToken)
	if err != nil {
		return nil, fmt.Errorf("creating MP config with seller token: %w", err)
	}

	prefClient := preference.NewClient(cfg)
	return c.createPreferenceWithClient(ctx, prefClient, req)
}

func (c *mpClient) createPreferenceWithClient(ctx context.Context, prefClient preference.Client, req *domain.MPPreferenceRequest) (*domain.MPPreferenceResponse, error) {
	items := make([]preference.ItemRequest, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, preference.ItemRequest{
			Title:      item.Title,
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			CurrencyID: "ARS",
		})
	}

	mpReq := preference.Request{
		Items: items,
		BackURLs: &preference.BackURLsRequest{
			Success: req.BackURLs.Success,
			Failure: req.BackURLs.Failure,
			Pending: req.BackURLs.Pending,
		},
		NotificationURL:   req.NotificationURL,
		ExternalReference: req.ExternalReference,
		AutoReturn:        req.AutoReturn,
	}

	if req.MarketplaceFee > 0 {
		mpReq.MarketplaceFee = req.MarketplaceFee
	}

	resp, err := prefClient.Create(ctx, mpReq)
	if err != nil {
		return nil, err
	}

	return &domain.MPPreferenceResponse{
		PreferenceID:     resp.ID,
		InitPoint:        resp.InitPoint,
		SandboxInitPoint: resp.SandboxInitPoint,
	}, nil
}

func (c *mpClient) GetPayment(ctx context.Context, paymentID int64) (*domain.MPPaymentResponse, error) {
	resp, err := c.paymentClient.Get(ctx, int(paymentID))
	if err != nil {
		return nil, err
	}
	return &domain.MPPaymentResponse{
		ID:                int64(resp.ID),
		Status:            mapStatus(resp.Status),
		StatusDetail:      resp.StatusDetail,
		PaymentMethod:     resp.PaymentMethodID,
		ExternalReference: resp.ExternalReference,
	}, nil
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
