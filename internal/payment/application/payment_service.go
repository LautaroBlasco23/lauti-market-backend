package application

import (
	"context"
	"errors"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	orderDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/order/domain"
	domain "github.com/LautaroBlasco23/lauti-market-backend/internal/payment/domain"
)

type PaymentService struct {
	repo      domain.Repository
	orderRepo orderDomain.Repository
	mpClient  domain.MPClient
	idGen     apiDomain.IDGenerator
}

func NewPaymentService(
	repo domain.Repository,
	orderRepo orderDomain.Repository,
	mpClient domain.MPClient,
	idGen apiDomain.IDGenerator,
) *PaymentService {
	return &PaymentService{
		repo:      repo,
		orderRepo: orderRepo,
		mpClient:  mpClient,
		idGen:     idGen,
	}
}

type CreatePaymentInput struct {
	OrderID      string
	UserID       string
	CardToken    string
	PayerEmail   string
	Installments int
}

type WebhookInput struct {
	MPPaymentID int64
}

func (s *PaymentService) CreatePayment(ctx context.Context, input CreatePaymentInput) (*domain.Payment, error) {
	order, err := s.orderRepo.FindByID(ctx, input.OrderID)
	if err != nil {
		return nil, apiDomain.ErrOrderNotFound
	}

	if order.UserID() != input.UserID {
		return nil, apiDomain.ErrForbidden
	}

	if order.Status() != orderDomain.StatusPending {
		return nil, apiDomain.ErrForbiddenTransition
	}

	existing, err := s.repo.FindByOrderID(ctx, input.OrderID)
	if err == nil && existing != nil {
		return nil, apiDomain.ErrPaymentAlreadyExists
	}

	paymentID := s.idGen.Generate()
	idempotencyKey := s.idGen.Generate()

	p := domain.NewPayment(paymentID, input.OrderID, input.UserID, idempotencyKey, order.TotalPrice())

	if saveErr := s.repo.Save(ctx, p); saveErr != nil {
		return nil, saveErr
	}

	installments := input.Installments
	if installments == 0 {
		installments = 1
	}

	mpResp, err := s.mpClient.CreatePayment(ctx, &domain.MPPaymentRequest{
		Amount:         order.TotalPrice(),
		Description:    "Order " + input.OrderID,
		PayerEmail:     input.PayerEmail,
		CardToken:      input.CardToken,
		Installments:   installments,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, apiDomain.ErrPaymentFailed
	}

	p.UpdateFromMP(mpResp.ID, mpResp.Status, mpResp.StatusDetail, mpResp.PaymentMethod)

	if err := s.repo.UpdateFromMP(ctx, p); err != nil {
		return nil, err
	}

	if mpResp.Status == domain.StatusApproved {
		_ = s.orderRepo.UpdateStatus(ctx, input.OrderID, orderDomain.StatusConfirmed) //nolint:errcheck
	}

	return p, nil
}

func (s *PaymentService) HandleWebhook(ctx context.Context, input WebhookInput) error {
	mpResp, err := s.mpClient.GetPayment(ctx, input.MPPaymentID)
	if err != nil {
		return err
	}

	p, err := s.repo.FindByMPPaymentID(ctx, input.MPPaymentID)
	if err != nil {
		if errors.Is(err, apiDomain.ErrPaymentNotFound) {
			return nil
		}
		return err
	}

	if p.Status() == mpResp.Status {
		return nil
	}

	p.UpdateFromMP(input.MPPaymentID, mpResp.Status, mpResp.StatusDetail, mpResp.PaymentMethod)

	if err := s.repo.UpdateFromMP(ctx, p); err != nil {
		return err
	}

	if mpResp.Status == domain.StatusApproved {
		_ = s.orderRepo.UpdateStatus(ctx, p.OrderID(), orderDomain.StatusConfirmed) //nolint:errcheck
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
