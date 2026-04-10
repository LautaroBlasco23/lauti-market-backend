package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	orderDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/order/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/payment/application"
	paymentDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/payment/domain"
	storeDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
)

// --- Mocks ---

type mockPaymentRepo struct {
	SaveFn          func(ctx context.Context, p *paymentDomain.Payment) error
	FindByIDFn      func(ctx context.Context, id string) (*paymentDomain.Payment, error)
	FindByOrderIDFn func(ctx context.Context, orderID string) (*paymentDomain.Payment, error)
	UpdateFromMPFn  func(ctx context.Context, p *paymentDomain.Payment) error
}

func (m *mockPaymentRepo) Save(ctx context.Context, p *paymentDomain.Payment) error {
	return m.SaveFn(ctx, p)
}

func (m *mockPaymentRepo) FindByID(ctx context.Context, id string) (*paymentDomain.Payment, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *mockPaymentRepo) FindByOrderID(ctx context.Context, orderID string) (*paymentDomain.Payment, error) {
	return m.FindByOrderIDFn(ctx, orderID)
}

func (m *mockPaymentRepo) UpdateFromMP(ctx context.Context, p *paymentDomain.Payment) error {
	return m.UpdateFromMPFn(ctx, p)
}

type mockOrderRepo struct {
	FindByIDFn     func(ctx context.Context, id string) (*orderDomain.Order, error)
	UpdateStatusFn func(ctx context.Context, id string, status orderDomain.OrderStatus) error
}

func (m *mockOrderRepo) Save(ctx context.Context, o *orderDomain.Order) error { return nil }
func (m *mockOrderRepo) FindByID(ctx context.Context, id string) (*orderDomain.Order, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *mockOrderRepo) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*orderDomain.Order, error) {
	return nil, nil
}

func (m *mockOrderRepo) FindByStoreID(ctx context.Context, storeID string, limit, offset int) ([]*orderDomain.Order, error) {
	return nil, nil
}

func (m *mockOrderRepo) UpdateStatus(ctx context.Context, id string, status orderDomain.OrderStatus) error {
	if m.UpdateStatusFn != nil {
		return m.UpdateStatusFn(ctx, id, status)
	}
	return nil
}

type mockMPClient struct {
	CreatePreferenceFn          func(ctx context.Context, req *paymentDomain.MPPreferenceRequest) (*paymentDomain.MPPreferenceResponse, error)
	CreatePreferenceWithTokenFn func(ctx context.Context, accessToken string, req *paymentDomain.MPPreferenceRequest) (*paymentDomain.MPPreferenceResponse, error)
	GetPaymentFn                func(ctx context.Context, paymentID int64) (*paymentDomain.MPPaymentResponse, error)
}

func (m *mockMPClient) CreatePreference(ctx context.Context, req *paymentDomain.MPPreferenceRequest) (*paymentDomain.MPPreferenceResponse, error) {
	return m.CreatePreferenceFn(ctx, req)
}

func (m *mockMPClient) CreatePreferenceWithToken(ctx context.Context, accessToken string, req *paymentDomain.MPPreferenceRequest) (*paymentDomain.MPPreferenceResponse, error) {
	if m.CreatePreferenceWithTokenFn != nil {
		return m.CreatePreferenceWithTokenFn(ctx, accessToken, req)
	}
	return m.CreatePreferenceFn(ctx, req)
}

func (m *mockMPClient) GetPayment(ctx context.Context, paymentID int64) (*paymentDomain.MPPaymentResponse, error) {
	return m.GetPaymentFn(ctx, paymentID)
}

type mockIDGen struct {
	ids []string
	idx int
}

func (m *mockIDGen) Generate() string {
	id := m.ids[m.idx%len(m.ids)]
	m.idx++
	return id
}

type mockStoreService struct {
	GetByIDFn func(ctx context.Context, id string) (*storeDomain.Store, error)
}

func (m *mockStoreService) GetByID(ctx context.Context, id string) (*storeDomain.Store, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	// Return a connected store by default
	s, _ := storeDomain.NewStore(id, storeDomain.CreateStoreInput{
		Name:        "Test Store",
		Description: "Test",
		Address:     "123 St",
		PhoneNumber: "555-0000",
	})
	// Simulate connected MP account
	s.ConnectMP("mp-user-123", "test-token", "test-refresh", time.Now().Add(time.Hour))
	return s, nil
}

// helpers

func makePendingOrder(id, userID string) *orderDomain.Order {
	item, _ := orderDomain.NewOrderItem(id+"-item", id, "prod-1", 1, 100.0)
	o, _ := orderDomain.NewOrder(id, userID, "store-1", []*orderDomain.OrderItem{item}, 100.0)
	return o
}

func makeConfirmedOrder(id, userID string) *orderDomain.Order {
	item, _ := orderDomain.NewOrderItem(id+"-item", id, "prod-1", 1, 100.0)
	return orderDomain.NewOrderFromDB(id, userID, "store-1",
		orderDomain.StatusConfirmed,
		[]*orderDomain.OrderItem{item},
		100.0,
		time.Now(), time.Now(),
	)
}

func newService(payRepo paymentDomain.Repository, orderRepo orderDomain.Repository, mpClient paymentDomain.MPClient, idGen apiDomain.IDGenerator) *application.PaymentService {
	return application.NewPaymentService(payRepo, orderRepo, &mockStoreService{}, mpClient, idGen, application.Config{
		FrontendBaseURL: "http://localhost:3000",
	})
}

// --- CreatePreference ---

func TestCreatePreference_HappyPath(t *testing.T) {
	order := makePendingOrder("order-1", "user-1")

	orderRepo := &mockOrderRepo{
		FindByIDFn: func(_ context.Context, _ string) (*orderDomain.Order, error) {
			return order, nil
		},
	}
	payRepo := &mockPaymentRepo{
		SaveFn: func(_ context.Context, _ *paymentDomain.Payment) error { return nil },
		FindByOrderIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return nil, apiDomain.ErrPaymentNotFound
		},
	}
	mpClient := &mockMPClient{
		CreatePreferenceFn: func(_ context.Context, _ *paymentDomain.MPPreferenceRequest) (*paymentDomain.MPPreferenceResponse, error) {
			return &paymentDomain.MPPreferenceResponse{
				PreferenceID:     "pref-123",
				InitPoint:        "https://mp.com/checkout",
				SandboxInitPoint: "https://sandbox.mp.com/checkout",
			}, nil
		},
	}
	idGen := &mockIDGen{ids: []string{"pay-1"}}

	svc := newService(payRepo, orderRepo, mpClient, idGen)
	result, err := svc.CreatePreference(context.Background(), application.CreatePreferenceInput{
		OrderIDs: []string{"order-1"},
		UserID:   "user-1",
	})

	require.NoError(t, err)
	assert.Equal(t, "pref-123", result.PreferenceID)
	assert.Equal(t, "https://sandbox.mp.com/checkout", result.SandboxInitPoint)
}

func TestCreatePreference_OrderNotFound(t *testing.T) {
	orderRepo := &mockOrderRepo{
		FindByIDFn: func(_ context.Context, _ string) (*orderDomain.Order, error) {
			return nil, apiDomain.ErrOrderNotFound
		},
	}
	svc := newService(&mockPaymentRepo{}, orderRepo, &mockMPClient{}, &mockIDGen{ids: []string{"x"}})

	_, err := svc.CreatePreference(context.Background(), application.CreatePreferenceInput{
		OrderIDs: []string{"order-1"},
		UserID:   "user-1",
	})
	assert.ErrorIs(t, err, apiDomain.ErrOrderNotFound)
}

func TestCreatePreference_WrongUser_Forbidden(t *testing.T) {
	order := makePendingOrder("order-1", "owner-user")
	orderRepo := &mockOrderRepo{
		FindByIDFn: func(_ context.Context, _ string) (*orderDomain.Order, error) {
			return order, nil
		},
	}
	svc := newService(&mockPaymentRepo{}, orderRepo, &mockMPClient{}, &mockIDGen{ids: []string{"x"}})

	_, err := svc.CreatePreference(context.Background(), application.CreatePreferenceInput{
		OrderIDs: []string{"order-1"},
		UserID:   "other-user",
	})
	assert.ErrorIs(t, err, apiDomain.ErrForbidden)
}

func TestCreatePreference_OrderNotPending_ForbiddenTransition(t *testing.T) {
	order := makeConfirmedOrder("order-1", "user-1")
	orderRepo := &mockOrderRepo{
		FindByIDFn: func(_ context.Context, _ string) (*orderDomain.Order, error) {
			return order, nil
		},
	}
	svc := newService(&mockPaymentRepo{}, orderRepo, &mockMPClient{}, &mockIDGen{ids: []string{"x"}})

	_, err := svc.CreatePreference(context.Background(), application.CreatePreferenceInput{
		OrderIDs: []string{"order-1"},
		UserID:   "user-1",
	})
	assert.ErrorIs(t, err, apiDomain.ErrForbiddenTransition)
}

func TestCreatePreference_AlreadyExists(t *testing.T) {
	order := makePendingOrder("order-1", "user-1")
	existingPayment := paymentDomain.NewPayment("pay-existing", "order-1", "user-1", "pref-x", 100.0)

	orderRepo := &mockOrderRepo{
		FindByIDFn: func(_ context.Context, _ string) (*orderDomain.Order, error) {
			return order, nil
		},
	}
	payRepo := &mockPaymentRepo{
		FindByOrderIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return existingPayment, nil
		},
	}
	svc := newService(payRepo, orderRepo, &mockMPClient{}, &mockIDGen{ids: []string{"x"}})

	_, err := svc.CreatePreference(context.Background(), application.CreatePreferenceInput{
		OrderIDs: []string{"order-1"},
		UserID:   "user-1",
	})
	assert.ErrorIs(t, err, apiDomain.ErrPaymentAlreadyExists)
}

func TestCreatePreference_MPError_ReturnsPaymentFailed(t *testing.T) {
	order := makePendingOrder("order-1", "user-1")
	orderRepo := &mockOrderRepo{
		FindByIDFn: func(_ context.Context, _ string) (*orderDomain.Order, error) {
			return order, nil
		},
	}
	payRepo := &mockPaymentRepo{
		FindByOrderIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return nil, apiDomain.ErrPaymentNotFound
		},
	}
	mpClient := &mockMPClient{
		CreatePreferenceFn: func(_ context.Context, _ *paymentDomain.MPPreferenceRequest) (*paymentDomain.MPPreferenceResponse, error) {
			return nil, errors.New("mp network error")
		},
	}
	svc := newService(payRepo, orderRepo, mpClient, &mockIDGen{ids: []string{"pay-1"}})

	_, err := svc.CreatePreference(context.Background(), application.CreatePreferenceInput{
		OrderIDs: []string{"order-1"},
		UserID:   "user-1",
	})
	assert.ErrorIs(t, err, apiDomain.ErrPaymentFailed)
}

func TestCreatePreference_MultipleOrders_AllSaved(t *testing.T) {
	orders := map[string]*orderDomain.Order{
		"order-1": makePendingOrder("order-1", "user-1"),
		"order-2": makePendingOrder("order-2", "user-1"),
	}

	orderRepo := &mockOrderRepo{
		FindByIDFn: func(_ context.Context, id string) (*orderDomain.Order, error) {
			if o, ok := orders[id]; ok {
				return o, nil
			}
			return nil, apiDomain.ErrOrderNotFound
		},
	}

	savedCount := 0
	payRepo := &mockPaymentRepo{
		SaveFn: func(_ context.Context, _ *paymentDomain.Payment) error {
			savedCount++
			return nil
		},
		FindByOrderIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return nil, apiDomain.ErrPaymentNotFound
		},
	}
	mpClient := &mockMPClient{
		CreatePreferenceFn: func(_ context.Context, req *paymentDomain.MPPreferenceRequest) (*paymentDomain.MPPreferenceResponse, error) {
			return &paymentDomain.MPPreferenceResponse{
				PreferenceID: "pref-multi",
				InitPoint:    "https://mp.com/checkout",
			}, nil
		},
	}
	idGen := &mockIDGen{ids: []string{"pay-1", "pay-2"}}

	svc := newService(payRepo, orderRepo, mpClient, idGen)
	result, err := svc.CreatePreference(context.Background(), application.CreatePreferenceInput{
		OrderIDs: []string{"order-1", "order-2"},
		UserID:   "user-1",
	})

	require.NoError(t, err)
	assert.Equal(t, "pref-multi", result.PreferenceID)
	assert.Equal(t, 2, savedCount, "expected one payment record saved per order")
}

// --- HandleWebhook ---

func TestHandleWebhook_StatusUpdate(t *testing.T) {
	existing := paymentDomain.NewPayment("pay-1", "order-1", "user-1", "pref-1", 100.0)

	payRepo := &mockPaymentRepo{
		FindByOrderIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return existing, nil
		},
		UpdateFromMPFn: func(_ context.Context, _ *paymentDomain.Payment) error { return nil },
	}
	mpClient := &mockMPClient{
		GetPaymentFn: func(_ context.Context, _ int64) (*paymentDomain.MPPaymentResponse, error) {
			return &paymentDomain.MPPaymentResponse{
				ID:                42,
				Status:            paymentDomain.StatusApproved,
				StatusDetail:      "accredited",
				PaymentMethod:     "credit_card",
				ExternalReference: "order-1",
			}, nil
		},
	}
	orderRepo := &mockOrderRepo{
		UpdateStatusFn: func(_ context.Context, _ string, _ orderDomain.OrderStatus) error { return nil },
	}
	svc := newService(payRepo, orderRepo, mpClient, &mockIDGen{ids: []string{"x"}})

	err := svc.HandleWebhook(context.Background(), application.WebhookInput{MPPaymentID: 42})
	assert.NoError(t, err)
}

func TestHandleWebhook_SameStatus_NoUpdate(t *testing.T) {
	existing := paymentDomain.NewPayment("pay-1", "order-1", "user-1", "pref-1", 100.0)

	updateCalled := false
	payRepo := &mockPaymentRepo{
		FindByOrderIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return existing, nil
		},
		UpdateFromMPFn: func(_ context.Context, _ *paymentDomain.Payment) error {
			updateCalled = true
			return nil
		},
	}
	mpClient := &mockMPClient{
		GetPaymentFn: func(_ context.Context, _ int64) (*paymentDomain.MPPaymentResponse, error) {
			return &paymentDomain.MPPaymentResponse{
				ID:                42,
				Status:            paymentDomain.StatusPending, // same as existing
				ExternalReference: "order-1",
			}, nil
		},
	}
	svc := newService(payRepo, &mockOrderRepo{}, mpClient, &mockIDGen{ids: []string{"x"}})

	err := svc.HandleWebhook(context.Background(), application.WebhookInput{MPPaymentID: 42})
	assert.NoError(t, err)
	assert.False(t, updateCalled, "UpdateFromMP should not be called when status is unchanged")
}

func TestHandleWebhook_UnknownOrder_SkipsGracefully(t *testing.T) {
	payRepo := &mockPaymentRepo{
		FindByOrderIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return nil, apiDomain.ErrPaymentNotFound
		},
	}
	mpClient := &mockMPClient{
		GetPaymentFn: func(_ context.Context, _ int64) (*paymentDomain.MPPaymentResponse, error) {
			return &paymentDomain.MPPaymentResponse{
				ID:                99,
				Status:            paymentDomain.StatusApproved,
				ExternalReference: "unknown-order",
			}, nil
		},
	}
	svc := newService(payRepo, &mockOrderRepo{}, mpClient, &mockIDGen{ids: []string{"x"}})

	err := svc.HandleWebhook(context.Background(), application.WebhookInput{MPPaymentID: 99})
	assert.NoError(t, err)
}

func TestHandleWebhook_EmptyExternalReference_SkipsGracefully(t *testing.T) {
	mpClient := &mockMPClient{
		GetPaymentFn: func(_ context.Context, _ int64) (*paymentDomain.MPPaymentResponse, error) {
			return &paymentDomain.MPPaymentResponse{
				ID:                1,
				Status:            paymentDomain.StatusApproved,
				ExternalReference: "",
			}, nil
		},
	}
	svc := newService(&mockPaymentRepo{}, &mockOrderRepo{}, mpClient, &mockIDGen{ids: []string{"x"}})

	err := svc.HandleWebhook(context.Background(), application.WebhookInput{MPPaymentID: 1})
	assert.NoError(t, err)
}

func TestHandleWebhook_MPError_PropagatesError(t *testing.T) {
	mpClient := &mockMPClient{
		GetPaymentFn: func(_ context.Context, _ int64) (*paymentDomain.MPPaymentResponse, error) {
			return nil, errors.New("mp api unavailable")
		},
	}
	svc := newService(&mockPaymentRepo{}, &mockOrderRepo{}, mpClient, &mockIDGen{ids: []string{"x"}})

	err := svc.HandleWebhook(context.Background(), application.WebhookInput{MPPaymentID: 1})
	assert.Error(t, err)
}

func TestHandleWebhook_MultipleOrders_AllUpdated(t *testing.T) {
	pay1 := paymentDomain.NewPayment("pay-1", "order-1", "user-1", "pref-1", 100.0)
	pay2 := paymentDomain.NewPayment("pay-2", "order-2", "user-1", "pref-1", 100.0)

	payments := map[string]*paymentDomain.Payment{
		"order-1": pay1,
		"order-2": pay2,
	}
	updateCount := 0

	payRepo := &mockPaymentRepo{
		FindByOrderIDFn: func(_ context.Context, orderID string) (*paymentDomain.Payment, error) {
			if p, ok := payments[orderID]; ok {
				return p, nil
			}
			return nil, apiDomain.ErrPaymentNotFound
		},
		UpdateFromMPFn: func(_ context.Context, _ *paymentDomain.Payment) error {
			updateCount++
			return nil
		},
	}
	mpClient := &mockMPClient{
		GetPaymentFn: func(_ context.Context, _ int64) (*paymentDomain.MPPaymentResponse, error) {
			return &paymentDomain.MPPaymentResponse{
				ID:                42,
				Status:            paymentDomain.StatusApproved,
				ExternalReference: "order-1,order-2",
			}, nil
		},
	}
	orderRepo := &mockOrderRepo{
		UpdateStatusFn: func(_ context.Context, _ string, _ orderDomain.OrderStatus) error { return nil },
	}
	svc := newService(payRepo, orderRepo, mpClient, &mockIDGen{ids: []string{"x"}})

	err := svc.HandleWebhook(context.Background(), application.WebhookInput{MPPaymentID: 42})
	assert.NoError(t, err)
	assert.Equal(t, 2, updateCount, "both payments should be updated")
}

// --- GetByID ---

func TestGetByID_HappyPath(t *testing.T) {
	p := paymentDomain.NewPayment("pay-1", "order-1", "user-1", "", 100.0)
	payRepo := &mockPaymentRepo{
		FindByIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return p, nil
		},
	}
	svc := newService(payRepo, &mockOrderRepo{}, &mockMPClient{}, &mockIDGen{ids: []string{"x"}})

	result, err := svc.GetByID(context.Background(), "pay-1", "user-1")
	require.NoError(t, err)
	assert.Equal(t, "pay-1", result.ID())
}

func TestGetByID_NotFound(t *testing.T) {
	payRepo := &mockPaymentRepo{
		FindByIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return nil, apiDomain.ErrPaymentNotFound
		},
	}
	svc := newService(payRepo, &mockOrderRepo{}, &mockMPClient{}, &mockIDGen{ids: []string{"x"}})

	_, err := svc.GetByID(context.Background(), "nonexistent", "user-1")
	assert.ErrorIs(t, err, apiDomain.ErrPaymentNotFound)
}

func TestGetByID_WrongUser_Forbidden(t *testing.T) {
	p := paymentDomain.NewPayment("pay-1", "order-1", "owner-user", "", 100.0)
	payRepo := &mockPaymentRepo{
		FindByIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return p, nil
		},
	}
	svc := newService(payRepo, &mockOrderRepo{}, &mockMPClient{}, &mockIDGen{ids: []string{"x"}})

	_, err := svc.GetByID(context.Background(), "pay-1", "other-user")
	assert.ErrorIs(t, err, apiDomain.ErrForbidden)
}

// --- GetByOrderID ---

func TestGetByOrderID_HappyPath(t *testing.T) {
	p := paymentDomain.NewPayment("pay-1", "order-1", "user-1", "", 100.0)
	payRepo := &mockPaymentRepo{
		FindByOrderIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return p, nil
		},
	}
	svc := newService(payRepo, &mockOrderRepo{}, &mockMPClient{}, &mockIDGen{ids: []string{"x"}})

	result, err := svc.GetByOrderID(context.Background(), "order-1", "user-1")
	require.NoError(t, err)
	assert.Equal(t, "pay-1", result.ID())
}

func TestGetByOrderID_NotFound(t *testing.T) {
	payRepo := &mockPaymentRepo{
		FindByOrderIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return nil, apiDomain.ErrPaymentNotFound
		},
	}
	svc := newService(payRepo, &mockOrderRepo{}, &mockMPClient{}, &mockIDGen{ids: []string{"x"}})

	_, err := svc.GetByOrderID(context.Background(), "order-1", "user-1")
	assert.ErrorIs(t, err, apiDomain.ErrPaymentNotFound)
}

func TestGetByOrderID_WrongUser_Forbidden(t *testing.T) {
	p := paymentDomain.NewPayment("pay-1", "order-1", "owner-user", "", 100.0)
	payRepo := &mockPaymentRepo{
		FindByOrderIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return p, nil
		},
	}
	svc := newService(payRepo, &mockOrderRepo{}, &mockMPClient{}, &mockIDGen{ids: []string{"x"}})

	_, err := svc.GetByOrderID(context.Background(), "order-1", "other-user")
	assert.ErrorIs(t, err, apiDomain.ErrForbidden)
}
