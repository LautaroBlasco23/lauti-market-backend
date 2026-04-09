package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	orderDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/order/domain"
	domain "github.com/LautaroBlasco23/lauti-market-backend/internal/payment/domain"
)

type Config struct {
	FrontendBaseURL string
	NotificationURL string
}

type PaymentService struct {
	repo      domain.Repository
	orderRepo orderDomain.Repository
	mpClient  domain.MPClient
	idGen     apiDomain.IDGenerator
	cfg       Config
}

func NewPaymentService(
	repo domain.Repository,
	orderRepo orderDomain.Repository,
	mpClient domain.MPClient,
	idGen apiDomain.IDGenerator,
	cfg Config,
) *PaymentService {
	return &PaymentService{
		repo:      repo,
		orderRepo: orderRepo,
		mpClient:  mpClient,
		idGen:     idGen,
		cfg:       cfg,
	}
}

type CreatePreferenceInput struct {
	OrderIDs []string
	UserID   string
}

type PreferenceResult struct {
	PreferenceID     string
	InitPoint        string
	SandboxInitPoint string
}

type WebhookInput struct {
	MPPaymentID int64
}

func (s *PaymentService) CreatePreference(ctx context.Context, input CreatePreferenceInput) (*PreferenceResult, error) {
	if len(input.OrderIDs) == 0 {
		return nil, apiDomain.ErrOrderNotFound
	}

	orders := make([]*orderDomain.Order, 0, len(input.OrderIDs))
	for _, orderID := range input.OrderIDs {
		order, err := s.orderRepo.FindByID(ctx, orderID)
		if err != nil {
			return nil, apiDomain.ErrOrderNotFound
		}
		if order.UserID() != input.UserID {
			return nil, apiDomain.ErrForbidden
		}
		if order.Status() != orderDomain.StatusPending {
			return nil, apiDomain.ErrForbiddenTransition
		}

		existing, err := s.repo.FindByOrderID(ctx, orderID)
		if err == nil && existing != nil {
			return nil, apiDomain.ErrPaymentAlreadyExists
		}

		orders = append(orders, order)
	}

	// Build preference items from all orders.
	items := make([]domain.MPPreferenceItem, 0)
	for _, order := range orders {
		for _, item := range order.Items() {
			items = append(items, domain.MPPreferenceItem{
				Title:     fmt.Sprintf("Product %s", item.ProductID()),
				Quantity:  item.Quantity(),
				UnitPrice: item.UnitPrice(),
			})
		}
	}

	externalRef := strings.Join(input.OrderIDs, ",")
	frontendBase := s.cfg.FrontendBaseURL
	if frontendBase == "" {
		frontendBase = "http://localhost:3000"
	}

	mpResp, err := s.mpClient.CreatePreference(ctx, &domain.MPPreferenceRequest{
		Items: items,
		BackURLs: domain.MPBackURLs{
			Success: frontendBase + "/checkout/success",
			Failure: frontendBase + "/checkout/failure",
			Pending: frontendBase + "/checkout/pending",
		},
		NotificationURL:   s.cfg.NotificationURL,
		ExternalReference: externalRef,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apiDomain.ErrPaymentFailed, err)
	}

	// Save one payment record per order.
	for _, order := range orders {
		p := domain.NewPayment(s.idGen.Generate(), order.ID(), input.UserID, mpResp.PreferenceID, order.TotalPrice())
		if saveErr := s.repo.Save(ctx, p); saveErr != nil {
			return nil, fmt.Errorf("saving payment for order %s: %w", order.ID(), saveErr)
		}
	}

	return &PreferenceResult{
		PreferenceID:     mpResp.PreferenceID,
		InitPoint:        mpResp.InitPoint,
		SandboxInitPoint: mpResp.SandboxInitPoint,
	}, nil
}

func (s *PaymentService) HandleWebhook(ctx context.Context, input WebhookInput) error {
	mpResp, err := s.mpClient.GetPayment(ctx, input.MPPaymentID)
	if err != nil {
		return err
	}

	if mpResp.ExternalReference == "" {
		return nil
	}

	orderIDs := strings.Split(mpResp.ExternalReference, ",")
	for _, orderID := range orderIDs {
		orderID = strings.TrimSpace(orderID)
		if orderID == "" {
			continue
		}

		p, err := s.repo.FindByOrderID(ctx, orderID)
		if err != nil {
			if errors.Is(err, apiDomain.ErrPaymentNotFound) {
				continue
			}
			return err
		}

		if p.Status() == mpResp.Status {
			continue
		}

		p.UpdateFromMP(input.MPPaymentID, mpResp.Status, mpResp.StatusDetail, mpResp.PaymentMethod)

		if err := s.repo.UpdateFromMP(ctx, p); err != nil {
			return err
		}

		if mpResp.Status == domain.StatusApproved {
			_ = s.orderRepo.UpdateStatus(ctx, orderID, orderDomain.StatusConfirmed) //nolint:errcheck
		}
	}

	return nil
}

func (s *PaymentService) GetByID(ctx context.Context, id, accountID string) (*domain.Payment, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apiDomain.ErrPaymentNotFound
	}

	if p.UserID() != accountID {
		return nil, apiDomain.ErrForbidden
	}

	return p, nil
}

func (s *PaymentService) GetByOrderID(ctx context.Context, orderID, accountID string) (*domain.Payment, error) {
	p, err := s.repo.FindByOrderID(ctx, orderID)
	if err != nil {
		return nil, apiDomain.ErrPaymentNotFound
	}

	if p.UserID() != accountID {
		return nil, apiDomain.ErrForbidden
	}

	return p, nil
}
