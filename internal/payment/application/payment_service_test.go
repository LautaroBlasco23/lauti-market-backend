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
)

// --- Mocks ---

type mockPaymentRepo struct {
	SaveFn              func(ctx context.Context, p *paymentDomain.Payment) error
	FindByIDFn          func(ctx context.Context, id string) (*paymentDomain.Payment, error)
	FindByOrderIDFn     func(ctx context.Context, orderID string) (*paymentDomain.Payment, error)
	FindByMPPaymentIDFn func(ctx context.Context, mpPaymentID int64) (*paymentDomain.Payment, error)
	UpdateFromMPFn      func(ctx context.Context, p *paymentDomain.Payment) error
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

func (m *mockPaymentRepo) FindByMPPaymentID(ctx context.Context, mpPaymentID int64) (*paymentDomain.Payment, error) {
	return m.FindByMPPaymentIDFn(ctx, mpPaymentID)
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
	CreatePaymentFn func(ctx context.Context, req *paymentDomain.MPPaymentRequest) (*paymentDomain.MPPaymentResponse, error)
	GetPaymentFn    func(ctx context.Context, paymentID int64) (*paymentDomain.MPPaymentResponse, error)
}

func (m *mockMPClient) CreatePayment(ctx context.Context, req *paymentDomain.MPPaymentRequest) (*paymentDomain.MPPaymentResponse, error) {
	return m.CreatePaymentFn(ctx, req)
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

// --- CreatePayment ---

func TestCreatePayment_HappyPath(t *testing.T) {
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
		UpdateFromMPFn: func(_ context.Context, _ *paymentDomain.Payment) error { return nil },
	}
	mpClient := &mockMPClient{
		CreatePaymentFn: func(_ context.Context, _ *paymentDomain.MPPaymentRequest) (*paymentDomain.MPPaymentResponse, error) {
			return &paymentDomain.MPPaymentResponse{
				ID:            42,
				Status:        paymentDomain.StatusApproved,
				StatusDetail:  "accredited",
				PaymentMethod: "credit_card",
			}, nil
		},
	}
	idGen := &mockIDGen{ids: []string{"pay-1", "idem-1"}}

	svc := application.NewPaymentService(payRepo, orderRepo, mpClient, idGen)
	p, err := svc.CreatePayment(context.Background(), application.CreatePaymentInput{
		OrderID:    "order-1",
		UserID:     "user-1",
		CardToken:  "tok_123",
		PayerEmail: "user@example.com",
	})

	require.NoError(t, err)
	assert.Equal(t, "pay-1", p.ID())
	assert.Equal(t, int64(42), p.MPPaymentID())
	assert.Equal(t, paymentDomain.StatusApproved, p.Status())
}

func TestCreatePayment_OrderNotFound(t *testing.T) {
	orderRepo := &mockOrderRepo{
		FindByIDFn: func(_ context.Context, _ string) (*orderDomain.Order, error) {
			return nil, apiDomain.ErrOrderNotFound
		},
	}
	svc := application.NewPaymentService(&mockPaymentRepo{}, orderRepo, &mockMPClient{}, &mockIDGen{ids: []string{"x"}})

	_, err := svc.CreatePayment(context.Background(), application.CreatePaymentInput{
		OrderID: "order-1",
		UserID:  "user-1",
	})
	assert.ErrorIs(t, err, apiDomain.ErrOrderNotFound)
}

func TestCreatePayment_WrongUser_Forbidden(t *testing.T) {
	order := makePendingOrder("order-1", "owner-user")
	orderRepo := &mockOrderRepo{
		FindByIDFn: func(_ context.Context, _ string) (*orderDomain.Order, error) {
			return order, nil
		},
	}
	svc := application.NewPaymentService(&mockPaymentRepo{}, orderRepo, &mockMPClient{}, &mockIDGen{ids: []string{"x"}})

	_, err := svc.CreatePayment(context.Background(), application.CreatePaymentInput{
		OrderID: "order-1",
		UserID:  "other-user",
	})
	assert.ErrorIs(t, err, apiDomain.ErrForbidden)
}

func TestCreatePayment_OrderNotPending_ForbiddenTransition(t *testing.T) {
	order := makeConfirmedOrder("order-1", "user-1")
	orderRepo := &mockOrderRepo{
		FindByIDFn: func(_ context.Context, _ string) (*orderDomain.Order, error) {
			return order, nil
		},
	}
	svc := application.NewPaymentService(&mockPaymentRepo{}, orderRepo, &mockMPClient{}, &mockIDGen{ids: []string{"x"}})

	_, err := svc.CreatePayment(context.Background(), application.CreatePaymentInput{
		OrderID: "order-1",
		UserID:  "user-1",
	})
	assert.ErrorIs(t, err, apiDomain.ErrForbiddenTransition)
}

func TestCreatePayment_AlreadyExists(t *testing.T) {
	order := makePendingOrder("order-1", "user-1")
	existingPayment := paymentDomain.NewPayment("pay-existing", "order-1", "user-1", "idem-x", 100.0)

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
	svc := application.NewPaymentService(payRepo, orderRepo, &mockMPClient{}, &mockIDGen{ids: []string{"x"}})

	_, err := svc.CreatePayment(context.Background(), application.CreatePaymentInput{
		OrderID: "order-1",
		UserID:  "user-1",
	})
	assert.ErrorIs(t, err, apiDomain.ErrPaymentAlreadyExists)
}

func TestCreatePayment_MPError_ReturnsPaymentFailed(t *testing.T) {
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
		CreatePaymentFn: func(_ context.Context, _ *paymentDomain.MPPaymentRequest) (*paymentDomain.MPPaymentResponse, error) {
			return nil, errors.New("mp network error")
		},
	}
	svc := application.NewPaymentService(payRepo, orderRepo, mpClient, &mockIDGen{ids: []string{"pay-1", "idem-1"}})

	_, err := svc.CreatePayment(context.Background(), application.CreatePaymentInput{
		OrderID:    "order-1",
		UserID:     "user-1",
		CardToken:  "tok_123",
		PayerEmail: "user@example.com",
	})
	assert.ErrorIs(t, err, apiDomain.ErrPaymentFailed)
}

func TestCreatePayment_DefaultInstallments(t *testing.T) {
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
		UpdateFromMPFn: func(_ context.Context, _ *paymentDomain.Payment) error { return nil },
	}

	var capturedInstallments int
	mpClient := &mockMPClient{
		CreatePaymentFn: func(_ context.Context, req *paymentDomain.MPPaymentRequest) (*paymentDomain.MPPaymentResponse, error) {
			capturedInstallments = req.Installments
			return &paymentDomain.MPPaymentResponse{
				ID:     1,
				Status: paymentDomain.StatusPending,
			}, nil
		},
	}
	svc := application.NewPaymentService(payRepo, orderRepo, mpClient, &mockIDGen{ids: []string{"pay-1", "idem-1"}})

	_, err := svc.CreatePayment(context.Background(), application.CreatePaymentInput{
		OrderID:      "order-1",
		UserID:       "user-1",
		Installments: 0, // should default to 1
	})
	require.NoError(t, err)
	assert.Equal(t, 1, capturedInstallments)
}

func TestCreatePayment_Approved_UpdatesOrderStatus(t *testing.T) {
	order := makePendingOrder("order-1", "user-1")

	var updatedStatus orderDomain.OrderStatus
	orderRepo := &mockOrderRepo{
		FindByIDFn: func(_ context.Context, _ string) (*orderDomain.Order, error) {
			return order, nil
		},
		UpdateStatusFn: func(_ context.Context, _ string, status orderDomain.OrderStatus) error {
			updatedStatus = status
			return nil
		},
	}
	payRepo := &mockPaymentRepo{
		SaveFn: func(_ context.Context, _ *paymentDomain.Payment) error { return nil },
		FindByOrderIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return nil, apiDomain.ErrPaymentNotFound
		},
		UpdateFromMPFn: func(_ context.Context, _ *paymentDomain.Payment) error { return nil },
	}
	mpClient := &mockMPClient{
		CreatePaymentFn: func(_ context.Context, _ *paymentDomain.MPPaymentRequest) (*paymentDomain.MPPaymentResponse, error) {
			return &paymentDomain.MPPaymentResponse{
				ID:     99,
				Status: paymentDomain.StatusApproved,
			}, nil
		},
	}
	svc := application.NewPaymentService(payRepo, orderRepo, mpClient, &mockIDGen{ids: []string{"pay-1", "idem-1"}})

	_, err := svc.CreatePayment(context.Background(), application.CreatePaymentInput{
		OrderID: "order-1",
		UserID:  "user-1",
	})
	require.NoError(t, err)
	assert.Equal(t, orderDomain.StatusConfirmed, updatedStatus)
}

// --- HandleWebhook ---

func TestHandleWebhook_StatusUpdate(t *testing.T) {
	existing := paymentDomain.NewPayment("pay-1", "order-1", "user-1", "idem-1", 100.0)
	// existing status is pending

	payRepo := &mockPaymentRepo{
		FindByMPPaymentIDFn: func(_ context.Context, _ int64) (*paymentDomain.Payment, error) {
			return existing, nil
		},
		UpdateFromMPFn: func(_ context.Context, _ *paymentDomain.Payment) error { return nil },
	}
	mpClient := &mockMPClient{
		GetPaymentFn: func(_ context.Context, _ int64) (*paymentDomain.MPPaymentResponse, error) {
			return &paymentDomain.MPPaymentResponse{
				ID:            42,
				Status:        paymentDomain.StatusApproved,
				StatusDetail:  "accredited",
				PaymentMethod: "credit_card",
			}, nil
		},
	}
	orderRepo := &mockOrderRepo{
		UpdateStatusFn: func(_ context.Context, _ string, _ orderDomain.OrderStatus) error { return nil },
	}
	svc := application.NewPaymentService(payRepo, orderRepo, mpClient, &mockIDGen{ids: []string{"x"}})

	err := svc.HandleWebhook(context.Background(), application.WebhookInput{MPPaymentID: 42})
	assert.NoError(t, err)
}

func TestHandleWebhook_SameStatus_NoUpdate(t *testing.T) {
	existing := paymentDomain.NewPayment("pay-1", "order-1", "user-1", "idem-1", 100.0)
	// existing status is pending

	updateCalled := false
	payRepo := &mockPaymentRepo{
		FindByMPPaymentIDFn: func(_ context.Context, _ int64) (*paymentDomain.Payment, error) {
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
				ID:     42,
				Status: paymentDomain.StatusPending, // same status as existing
			}, nil
		},
	}
	svc := application.NewPaymentService(payRepo, &mockOrderRepo{}, mpClient, &mockIDGen{ids: []string{"x"}})

	err := svc.HandleWebhook(context.Background(), application.WebhookInput{MPPaymentID: 42})
	assert.NoError(t, err)
	assert.False(t, updateCalled, "UpdateFromMP should not be called when status is unchanged")
}

func TestHandleWebhook_UnknownMPPaymentID_ReturnsNil(t *testing.T) {
	payRepo := &mockPaymentRepo{
		FindByMPPaymentIDFn: func(_ context.Context, _ int64) (*paymentDomain.Payment, error) {
			return nil, apiDomain.ErrPaymentNotFound
		},
	}
	mpClient := &mockMPClient{
		GetPaymentFn: func(_ context.Context, _ int64) (*paymentDomain.MPPaymentResponse, error) {
			return &paymentDomain.MPPaymentResponse{
				ID:     99,
				Status: paymentDomain.StatusApproved,
			}, nil
		},
	}
	svc := application.NewPaymentService(payRepo, &mockOrderRepo{}, mpClient, &mockIDGen{ids: []string{"x"}})

	err := svc.HandleWebhook(context.Background(), application.WebhookInput{MPPaymentID: 99})
	assert.NoError(t, err)
}

func TestHandleWebhook_MPError_PropagatesError(t *testing.T) {
	mpClient := &mockMPClient{
		GetPaymentFn: func(_ context.Context, _ int64) (*paymentDomain.MPPaymentResponse, error) {
			return nil, errors.New("mp api unavailable")
		},
	}
	svc := application.NewPaymentService(&mockPaymentRepo{}, &mockOrderRepo{}, mpClient, &mockIDGen{ids: []string{"x"}})

	err := svc.HandleWebhook(context.Background(), application.WebhookInput{MPPaymentID: 1})
	assert.Error(t, err)
}

// --- GetByID ---

func TestGetByID_HappyPath(t *testing.T) {
	p := paymentDomain.NewPayment("pay-1", "order-1", "user-1", "idem-1", 100.0)
	payRepo := &mockPaymentRepo{
		FindByIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return p, nil
		},
	}
	svc := application.NewPaymentService(payRepo, &mockOrderRepo{}, &mockMPClient{}, &mockIDGen{ids: []string{"x"}})

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
	svc := application.NewPaymentService(payRepo, &mockOrderRepo{}, &mockMPClient{}, &mockIDGen{ids: []string{"x"}})

	_, err := svc.GetByID(context.Background(), "nonexistent", "user-1")
	assert.ErrorIs(t, err, apiDomain.ErrPaymentNotFound)
}

func TestGetByID_WrongUser_Forbidden(t *testing.T) {
	p := paymentDomain.NewPayment("pay-1", "order-1", "owner-user", "idem-1", 100.0)
	payRepo := &mockPaymentRepo{
		FindByIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return p, nil
		},
	}
	svc := application.NewPaymentService(payRepo, &mockOrderRepo{}, &mockMPClient{}, &mockIDGen{ids: []string{"x"}})

	_, err := svc.GetByID(context.Background(), "pay-1", "other-user")
	assert.ErrorIs(t, err, apiDomain.ErrForbidden)
}

// --- GetByOrderID ---

func TestGetByOrderID_HappyPath(t *testing.T) {
	p := paymentDomain.NewPayment("pay-1", "order-1", "user-1", "idem-1", 100.0)
	payRepo := &mockPaymentRepo{
		FindByOrderIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return p, nil
		},
	}
	svc := application.NewPaymentService(payRepo, &mockOrderRepo{}, &mockMPClient{}, &mockIDGen{ids: []string{"x"}})

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
	svc := application.NewPaymentService(payRepo, &mockOrderRepo{}, &mockMPClient{}, &mockIDGen{ids: []string{"x"}})

	_, err := svc.GetByOrderID(context.Background(), "order-1", "user-1")
	assert.ErrorIs(t, err, apiDomain.ErrPaymentNotFound)
}

func TestGetByOrderID_WrongUser_Forbidden(t *testing.T) {
	p := paymentDomain.NewPayment("pay-1", "order-1", "owner-user", "idem-1", 100.0)
	payRepo := &mockPaymentRepo{
		FindByOrderIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return p, nil
		},
	}
	svc := application.NewPaymentService(payRepo, &mockOrderRepo{}, &mockMPClient{}, &mockIDGen{ids: []string{"x"}})

	_, err := svc.GetByOrderID(context.Background(), "order-1", "other-user")
	assert.ErrorIs(t, err, apiDomain.ErrForbidden)
}
