package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	orderDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/order/domain"
	domain "github.com/LautaroBlasco23/lauti-market-backend/internal/payment/domain"
	productDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/product/domain"
	storeDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
)

const marketplaceFeePercent = 5.0

type Config struct {
	FrontendBaseURL string
	NotificationURL string
}

// StoreService defines the interface needed from store module
type StoreService interface {
	GetByID(ctx context.Context, id string) (*storeDomain.Store, error)
}

type PaymentService struct {
	repo        domain.Repository
	orderRepo   orderDomain.Repository
	productRepo productDomain.Repository
	storeSvc    StoreService
	mpClient    domain.MPClient
	idGen       apiDomain.IDGenerator
	cfg         Config
}

func NewPaymentService(
	repo domain.Repository,
	orderRepo orderDomain.Repository,
	productRepo productDomain.Repository,
	storeSvc StoreService,
	mpClient domain.MPClient,
	idGen apiDomain.IDGenerator,
	cfg Config,
) *PaymentService {
	return &PaymentService{
		repo:        repo,
		orderRepo:   orderRepo,
		productRepo: productRepo,
		storeSvc:    storeSvc,
		mpClient:    mpClient,
		idGen:       idGen,
		cfg:         cfg,
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
	slog.Debug("PaymentService.CreatePreference started",
		slog.String("user_id", input.UserID),
		slog.Int("order_count", len(input.OrderIDs)),
	)

	if len(input.OrderIDs) == 0 {
		slog.Error("PaymentService.CreatePreference failed",
			slog.String("operation", "validate_order_ids"),
			slog.String("error", "no order IDs provided"),
		)
		return nil, apiDomain.ErrOrderNotFound
	}

	orders := make([]*orderDomain.Order, 0, len(input.OrderIDs))
	for _, orderID := range input.OrderIDs {
		slog.Debug("PaymentService.CreatePreference fetching order", slog.String("order_id", orderID))
		order, err := s.orderRepo.FindByID(ctx, orderID)
		if err != nil {
			slog.Error("PaymentService.CreatePreference failed",
				slog.String("operation", "find_order_by_id"),
				slog.String("order_id", orderID),
				slog.Any("error", err),
			)
			return nil, apiDomain.ErrOrderNotFound
		}
		if order.UserID() != input.UserID {
			slog.Error("PaymentService.CreatePreference failed",
				slog.String("operation", "validate_order_ownership"),
				slog.String("order_id", orderID),
				slog.String("order_user_id", order.UserID()),
				slog.String("request_user_id", input.UserID),
				slog.String("error", "forbidden access to order"),
			)
			return nil, apiDomain.ErrForbidden
		}
		if order.Status() != orderDomain.StatusPending {
			slog.Error("PaymentService.CreatePreference failed",
				slog.String("operation", "validate_order_status"),
				slog.String("order_id", orderID),
				slog.String("current_status", string(order.Status())),
				slog.String("error", "order not in pending status"),
			)
			return nil, apiDomain.ErrForbiddenTransition
		}

		existing, err := s.repo.FindByOrderID(ctx, orderID)
		if err == nil && existing != nil {
			slog.Error("PaymentService.CreatePreference failed",
				slog.String("operation", "check_existing_payment"),
				slog.String("order_id", orderID),
				slog.String("error", "payment already exists"),
			)
			return nil, apiDomain.ErrPaymentAlreadyExists
		}

		orders = append(orders, order)
	}

	// All orders must belong to the same store for seller-specific MP token
	if len(orders) == 0 {
		return nil, apiDomain.ErrOrderNotFound
	}
	storeID := orders[0].StoreID()
	for _, order := range orders {
		if order.StoreID() != storeID {
			slog.Error("PaymentService.CreatePreference failed",
				slog.String("operation", "validate_same_store"),
				slog.String("error", "orders belong to different stores"),
			)
			return nil, apiDomain.ErrForbidden
		}
	}

	// Fetch store and verify MP connection
	store, err := s.storeSvc.GetByID(ctx, storeID)
	if err != nil {
		slog.Error("PaymentService.CreatePreference failed",
			slog.String("operation", "find_store_by_id"),
			slog.String("store_id", storeID),
			slog.Any("error", err),
		)
		return nil, storeDomain.ErrStoreNotFound
	}

	if !store.IsMPConnected() {
		slog.Error("PaymentService.CreatePreference failed",
			slog.String("operation", "check_mp_connected"),
			slog.String("store_id", storeID),
			slog.String("error", "store does not have a connected MercadoPago account"),
		)
		return nil, apiDomain.ErrStoreMPNotConnected
	}

	if !store.IsMPTokenValid() {
		slog.Error("PaymentService.CreatePreference failed",
			slog.String("operation", "check_mp_token_valid"),
			slog.String("store_id", storeID),
			slog.String("error", "store MercadoPago token is invalid or expired"),
		)
		return nil, apiDomain.ErrStoreMPTokenExpired
	}

	// Build preference items from all orders.
	items := make([]domain.MPPreferenceItem, 0)
	var totalAmount float64
	for _, order := range orders {
		for _, item := range order.Items() {
			items = append(items, domain.MPPreferenceItem{
				Title:     fmt.Sprintf("Product %s", item.ProductID()),
				Quantity:  item.Quantity(),
				UnitPrice: item.UnitPrice(),
			})
		}
		totalAmount += order.TotalPrice()
	}

	// Calculate marketplace fee (5% of total)
	marketplaceFee := totalAmount * marketplaceFeePercent / 100

	externalRef := strings.Join(input.OrderIDs, ",")
	frontendBase := s.cfg.FrontendBaseURL
	if frontendBase == "" {
		frontendBase = "http://localhost:3000"
	}

	slog.Debug("PaymentService.CreatePreference calling MercadoPago",
		slog.String("external_reference", externalRef),
		slog.Int("item_count", len(items)),
		slog.String("store_id", storeID),
		slog.Float64("marketplace_fee", marketplaceFee),
	)
	mpResp, err := s.mpClient.CreatePreferenceWithToken(ctx, store.MPAccessToken(), &domain.MPPreferenceRequest{
		Items: items,
		BackURLs: domain.MPBackURLs{
			Success: frontendBase + "/checkout/success",
			Failure: frontendBase + "/checkout/failure",
			Pending: frontendBase + "/checkout/pending",
		},
		NotificationURL:   s.cfg.NotificationURL,
		ExternalReference: externalRef,
		MarketplaceFee:    marketplaceFee,
	})
	if err != nil {
		slog.Error("PaymentService.CreatePreference failed",
			slog.String("operation", "create_mp_preference"),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("%w: %w", apiDomain.ErrPaymentFailed, err)
	}

	// Save one payment record per order.
	slog.Debug("PaymentService.CreatePreference saving payment records",
		slog.String("preference_id", mpResp.PreferenceID),
		slog.Int("order_count", len(orders)),
	)
	for _, order := range orders {
		p := domain.NewPayment(s.idGen.Generate(), order.ID(), input.UserID, mpResp.PreferenceID, order.TotalPrice())
		if saveErr := s.repo.Save(ctx, p); saveErr != nil {
			slog.Error("PaymentService.CreatePreference failed",
				slog.String("operation", "save_payment"),
				slog.String("order_id", order.ID()),
				slog.Any("error", saveErr),
			)
			return nil, fmt.Errorf("saving payment for order %s: %w", order.ID(), saveErr)
		}
	}

	slog.Info("PaymentService.CreatePreference completed",
		slog.String("preference_id", mpResp.PreferenceID),
		slog.String("user_id", input.UserID),
		slog.Int("order_count", len(orders)),
	)
	return &PreferenceResult{
		PreferenceID:     mpResp.PreferenceID,
		InitPoint:        mpResp.InitPoint,
		SandboxInitPoint: mpResp.SandboxInitPoint,
	}, nil
}

type CreateCartPreferenceInput struct {
	Items  []CartItemInput
	UserID string
}

type CartItemInput struct {
	ProductID string
	Quantity  int
	UnitPrice float64
}

func (s *PaymentService) CreateCartPreference(ctx context.Context, input CreateCartPreferenceInput) (*PreferenceResult, error) {
	slog.Debug("PaymentService.CreateCartPreference started",
		slog.String("user_id", input.UserID),
		slog.Int("item_count", len(input.Items)),
	)

	if len(input.Items) == 0 {
		slog.Error("PaymentService.CreateCartPreference failed",
			slog.String("operation", "validate_items"),
			slog.String("error", "no items provided"),
		)
		return nil, apiDomain.ErrInvalidPaymentAmount
	}

	// Validate all products and collect store IDs
	storeIDs := make(map[string]bool)
	items := make([]domain.MPPreferenceItem, 0, len(input.Items))

	for _, item := range input.Items {
		slog.Debug("PaymentService.CreateCartPreference fetching product",
			slog.String("product_id", item.ProductID),
		)

		product, err := s.productRepo.FindByID(ctx, item.ProductID)
		if err != nil {
			slog.Error("PaymentService.CreateCartPreference failed",
				slog.String("operation", "find_product_by_id"),
				slog.String("product_id", item.ProductID),
				slog.Any("error", err),
			)
			return nil, fmt.Errorf("product not found: %s", item.ProductID)
		}

		// Validate price matches
		if product.Price() != item.UnitPrice {
			slog.Error("PaymentService.CreateCartPreference failed",
				slog.String("operation", "validate_price"),
				slog.String("product_id", item.ProductID),
				slog.Float64("expected_price", product.Price()),
				slog.Float64("provided_price", item.UnitPrice),
			)
			return nil, apiDomain.ErrInvalidPaymentAmount
		}

		storeIDs[product.StoreID()] = true
		items = append(items, domain.MPPreferenceItem{
			Title:     product.Name(),
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}

	// Multi-store checkout: verify all stores have MP connected
	for storeID := range storeIDs {
		store, err := s.storeSvc.GetByID(ctx, storeID)
		if err != nil {
			slog.Error("PaymentService.CreateCartPreference failed",
				slog.String("operation", "find_store_by_id"),
				slog.String("store_id", storeID),
				slog.Any("error", err),
			)
			return nil, storeDomain.ErrStoreNotFound
		}

		if !store.IsMPConnected() {
			slog.Error("PaymentService.CreateCartPreference failed",
				slog.String("operation", "check_mp_connected"),
				slog.String("store_id", storeID),
				slog.String("error", "store does not have a connected MercadoPago account"),
			)
			return nil, apiDomain.ErrStoreMPNotConnected
		}

		if !store.IsMPTokenValid() {
			slog.Error("PaymentService.CreateCartPreference failed",
				slog.String("operation", "check_mp_token_valid"),
				slog.String("store_id", storeID),
				slog.String("error", "store MercadoPago token is invalid or expired"),
			)
			return nil, apiDomain.ErrStoreMPTokenExpired
		}
	}

	// Use userID + unique ID as external reference (orders don't exist yet)
	externalRef := fmt.Sprintf("cart_%s_%s", input.UserID, s.idGen.Generate())
	frontendBase := s.cfg.FrontendBaseURL
	if frontendBase == "" {
		frontendBase = "http://localhost:3000"
	}

	slog.Debug("PaymentService.CreateCartPreference calling MercadoPago",
		slog.String("external_reference", externalRef),
		slog.Int("item_count", len(items)),
		slog.Int("store_count", len(storeIDs)),
	)

	// Create preference using marketplace credentials (not seller token).
	// marketplace_fee is intentionally omitted here: that field is only valid when
	// the preference is created with a seller's OAuth token (split-payment flow).
	// Using it with our own marketplace token causes MercadoPago to return an auth error.
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
		slog.Error("PaymentService.CreateCartPreference failed",
			slog.String("operation", "create_mp_preference"),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("%w: %w", apiDomain.ErrPaymentFailed, err)
	}

	slog.Info("PaymentService.CreateCartPreference completed",
		slog.String("preference_id", mpResp.PreferenceID),
		slog.String("user_id", input.UserID),
		slog.Int("store_count", len(storeIDs)),
	)
	return &PreferenceResult{
		PreferenceID:     mpResp.PreferenceID,
		InitPoint:        mpResp.InitPoint,
		SandboxInitPoint: mpResp.SandboxInitPoint,
	}, nil
}

func (s *PaymentService) HandleWebhook(ctx context.Context, input WebhookInput) error {
	slog.Debug("PaymentService.HandleWebhook started",
		slog.Int64("mp_payment_id", input.MPPaymentID),
	)

	mpResp, err := s.mpClient.GetPayment(ctx, input.MPPaymentID)
	if err != nil {
		slog.Error("PaymentService.HandleWebhook failed",
			slog.String("operation", "get_mp_payment"),
			slog.Int64("mp_payment_id", input.MPPaymentID),
			slog.Any("error", err),
		)
		return err
	}

	if mpResp.ExternalReference == "" {
		slog.Debug("PaymentService.HandleWebhook skipped - no external reference",
			slog.Int64("mp_payment_id", input.MPPaymentID),
		)
		return nil
	}

	orderIDs := strings.Split(mpResp.ExternalReference, ",")
	slog.Debug("PaymentService.HandleWebhook processing orders",
		slog.Int64("mp_payment_id", input.MPPaymentID),
		slog.String("mp_status", string(mpResp.Status)),
		slog.Int("order_count", len(orderIDs)),
	)

	for _, orderID := range orderIDs {
		orderID = strings.TrimSpace(orderID)
		if orderID == "" {
			continue
		}

		p, err := s.repo.FindByOrderID(ctx, orderID)
		if err != nil {
			if errors.Is(err, apiDomain.ErrPaymentNotFound) {
				slog.Debug("PaymentService.HandleWebhook payment not found for order, skipping",
					slog.String("order_id", orderID),
				)
				continue
			}
			slog.Error("PaymentService.HandleWebhook failed",
				slog.String("operation", "find_payment_by_order_id"),
				slog.String("order_id", orderID),
				slog.Any("error", err),
			)
			return err
		}

		if p.Status() == mpResp.Status {
			slog.Debug("PaymentService.HandleWebhook status unchanged, skipping",
				slog.String("order_id", orderID),
				slog.String("status", string(p.Status())),
			)
			continue
		}

		slog.Debug("PaymentService.HandleWebhook updating payment status",
			slog.String("order_id", orderID),
			slog.String("old_status", string(p.Status())),
			slog.String("new_status", string(mpResp.Status)),
		)
		p.UpdateFromMP(input.MPPaymentID, mpResp.Status, mpResp.StatusDetail, mpResp.PaymentMethod)

		if err := s.repo.UpdateFromMP(ctx, p); err != nil {
			slog.Error("PaymentService.HandleWebhook failed",
				slog.String("operation", "update_payment_from_mp"),
				slog.String("order_id", orderID),
				slog.Any("error", err),
			)
			return err
		}

		if mpResp.Status == domain.StatusApproved {
			slog.Debug("PaymentService.HandleWebhook confirming order",
				slog.String("order_id", orderID),
			)
			_ = s.orderRepo.UpdateStatus(ctx, orderID, orderDomain.StatusConfirmed) //nolint:errcheck
		}
	}

	slog.Info("PaymentService.HandleWebhook completed",
		slog.Int64("mp_payment_id", input.MPPaymentID),
		slog.String("mp_status", string(mpResp.Status)),
	)
	return nil
}

func (s *PaymentService) GetByID(ctx context.Context, id, accountID string) (*domain.Payment, error) {
	slog.Debug("PaymentService.GetByID started",
		slog.String("payment_id", id),
		slog.String("account_id", accountID),
	)

	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		slog.Error("PaymentService.GetByID failed",
			slog.String("operation", "find_by_id"),
			slog.String("payment_id", id),
			slog.Any("error", err),
		)
		return nil, apiDomain.ErrPaymentNotFound
	}

	if p.UserID() != accountID {
		slog.Error("PaymentService.GetByID failed",
			slog.String("operation", "validate_ownership"),
			slog.String("payment_id", id),
			slog.String("payment_user_id", p.UserID()),
			slog.String("request_account_id", accountID),
			slog.String("error", "forbidden access to payment"),
		)
		return nil, apiDomain.ErrForbidden
	}

	slog.Info("PaymentService.GetByID completed", slog.String("payment_id", id))
	return p, nil
}

func (s *PaymentService) GetByOrderID(ctx context.Context, orderID, accountID string) (*domain.Payment, error) {
	slog.Debug("PaymentService.GetByOrderID started",
		slog.String("order_id", orderID),
		slog.String("account_id", accountID),
	)

	p, err := s.repo.FindByOrderID(ctx, orderID)
	if err != nil {
		slog.Error("PaymentService.GetByOrderID failed",
			slog.String("operation", "find_by_order_id"),
			slog.String("order_id", orderID),
			slog.Any("error", err),
		)
		return nil, apiDomain.ErrPaymentNotFound
	}

	if p.UserID() != accountID {
		slog.Error("PaymentService.GetByOrderID failed",
			slog.String("operation", "validate_ownership"),
			slog.String("order_id", orderID),
			slog.String("payment_user_id", p.UserID()),
			slog.String("request_account_id", accountID),
			slog.String("error", "forbidden access to payment"),
		)
		return nil, apiDomain.ErrForbidden
	}

	slog.Info("PaymentService.GetByOrderID completed", slog.String("order_id", orderID))
	return p, nil
}
