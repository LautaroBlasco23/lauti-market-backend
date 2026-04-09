package controller_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	apiInfra "github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	authDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/auth/domain"
	authUtils "github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/utils"
	orderDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/order/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/payment/application"
	paymentDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/payment/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/payment/infrastructure/controller"
)

// --- Mocks ---

type mockPaymentRepo struct {
	SaveFn          func(ctx context.Context, p *paymentDomain.Payment) error
	FindByIDFn      func(ctx context.Context, id string) (*paymentDomain.Payment, error)
	FindByOrderIDFn func(ctx context.Context, orderID string) (*paymentDomain.Payment, error)
	UpdateFromMPFn  func(ctx context.Context, p *paymentDomain.Payment) error
}

func (m *mockPaymentRepo) Save(ctx context.Context, p *paymentDomain.Payment) error {
	if m.SaveFn != nil {
		return m.SaveFn(ctx, p)
	}
	return nil
}

func (m *mockPaymentRepo) FindByID(ctx context.Context, id string) (*paymentDomain.Payment, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *mockPaymentRepo) FindByOrderID(ctx context.Context, orderID string) (*paymentDomain.Payment, error) {
	return m.FindByOrderIDFn(ctx, orderID)
}

func (m *mockPaymentRepo) UpdateFromMP(ctx context.Context, p *paymentDomain.Payment) error {
	if m.UpdateFromMPFn != nil {
		return m.UpdateFromMPFn(ctx, p)
	}
	return nil
}

type mockOrderRepo struct {
	FindByIDFn     func(ctx context.Context, id string) (*orderDomain.Order, error)
	UpdateStatusFn func(ctx context.Context, id string, status orderDomain.OrderStatus) error
}

func (m *mockOrderRepo) Save(ctx context.Context, o *orderDomain.Order) error { return nil }
func (m *mockOrderRepo) FindByID(ctx context.Context, id string) (*orderDomain.Order, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, apiDomain.ErrOrderNotFound
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
	CreatePreferenceFn func(ctx context.Context, req *paymentDomain.MPPreferenceRequest) (*paymentDomain.MPPreferenceResponse, error)
	GetPaymentFn       func(ctx context.Context, paymentID int64) (*paymentDomain.MPPaymentResponse, error)
}

func (m *mockMPClient) CreatePreference(ctx context.Context, req *paymentDomain.MPPreferenceRequest) (*paymentDomain.MPPreferenceResponse, error) {
	if m.CreatePreferenceFn != nil {
		return m.CreatePreferenceFn(ctx, req)
	}
	return nil, apiDomain.ErrPaymentFailed
}

func (m *mockMPClient) GetPayment(ctx context.Context, paymentID int64) (*paymentDomain.MPPaymentResponse, error) {
	if m.GetPaymentFn != nil {
		return m.GetPaymentFn(ctx, paymentID)
	}
	return nil, apiDomain.ErrPaymentNotFound
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

func testJWTGen() *authUtils.JWTGenerator {
	return authUtils.NewJWTGenerator("test-secret", time.Hour)
}

func withClaims(t *testing.T, handler http.Handler, accountType, accountID string) (http.Handler, string) {
	t.Helper()
	jwtGen := testJWTGen()
	token, err := jwtGen.Generate("auth-1", authDomain.AccountType(accountType), accountID)
	require.NoError(t, err)
	mw := apiInfra.NewAuthMiddleware(jwtGen)
	return mw.Wrap(handler), token
}

func jsonBody(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewReader(b)
}

func makePendingOrder(id, userID string) *orderDomain.Order {
	item, _ := orderDomain.NewOrderItem(id+"-item", id, "prod-1", 1, 100.0)
	o, _ := orderDomain.NewOrder(id, userID, "store-1", []*orderDomain.OrderItem{item}, 100.0)
	return o
}

func makeTestPayment(id, orderID, userID string) *paymentDomain.Payment {
	return paymentDomain.NewPayment(id, orderID, userID, "pref-"+id, 100.0)
}

func makeController(payRepo paymentDomain.Repository, orderRepo orderDomain.Repository, mpClient paymentDomain.MPClient, webhookSecret string) *controller.PaymentController {
	idGen := &mockIDGen{ids: []string{"pay-1", "pay-2"}}
	svc := application.NewPaymentService(payRepo, orderRepo, mpClient, idGen, application.Config{
		FrontendBaseURL: "http://localhost:3000",
	})
	return controller.NewPaymentController(svc, webhookSecret)
}

// --- Create ---

func TestCreate_UserAccount_HappyPath(t *testing.T) {
	order := makePendingOrder("order-1", "user-1")
	payRepo := &mockPaymentRepo{
		SaveFn: func(_ context.Context, _ *paymentDomain.Payment) error { return nil },
		FindByOrderIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return nil, apiDomain.ErrPaymentNotFound
		},
	}
	orderRepo := &mockOrderRepo{
		FindByIDFn: func(_ context.Context, _ string) (*orderDomain.Order, error) {
			return order, nil
		},
	}
	mpClient := &mockMPClient{
		CreatePreferenceFn: func(_ context.Context, _ *paymentDomain.MPPreferenceRequest) (*paymentDomain.MPPreferenceResponse, error) {
			return &paymentDomain.MPPreferenceResponse{
				PreferenceID:     "pref-abc",
				InitPoint:        "https://mp.com/checkout",
				SandboxInitPoint: "https://sandbox.mp.com/checkout",
			}, nil
		},
	}

	ctrl := makeController(payRepo, orderRepo, mpClient, "")
	handler, token := withClaims(t, http.HandlerFunc(ctrl.Create), "user", "user-1")

	req := httptest.NewRequest(http.MethodPost, "/payments", jsonBody(t, map[string]any{
		"order_ids": []string{"order-1"},
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "pref-abc", resp["preference_id"])
	assert.Equal(t, "https://sandbox.mp.com/checkout", resp["sandbox_init_point"])
}

func TestCreate_StoreAccount_Forbidden(t *testing.T) {
	ctrl := makeController(&mockPaymentRepo{}, &mockOrderRepo{}, &mockMPClient{}, "")
	handler, token := withClaims(t, http.HandlerFunc(ctrl.Create), "store", "store-1")

	req := httptest.NewRequest(http.MethodPost, "/payments", jsonBody(t, map[string]any{
		"order_ids": []string{"order-1"},
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestCreate_NoAuth(t *testing.T) {
	ctrl := makeController(&mockPaymentRepo{}, &mockOrderRepo{}, &mockMPClient{}, "")

	req := httptest.NewRequest(http.MethodPost, "/payments", jsonBody(t, map[string]any{
		"order_ids": []string{"order-1"},
	}))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	ctrl.Create(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestCreate_InvalidBody(t *testing.T) {
	ctrl := makeController(&mockPaymentRepo{}, &mockOrderRepo{}, &mockMPClient{}, "")
	handler, token := withClaims(t, http.HandlerFunc(ctrl.Create), "user", "user-1")

	req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreate_ValidationError_MissingOrderIDs(t *testing.T) {
	ctrl := makeController(&mockPaymentRepo{}, &mockOrderRepo{}, &mockMPClient{}, "")
	handler, token := withClaims(t, http.HandlerFunc(ctrl.Create), "user", "user-1")

	req := httptest.NewRequest(http.MethodPost, "/payments", jsonBody(t, map[string]any{}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreate_OrderNotFound(t *testing.T) {
	orderRepo := &mockOrderRepo{
		FindByIDFn: func(_ context.Context, _ string) (*orderDomain.Order, error) {
			return nil, apiDomain.ErrOrderNotFound
		},
	}
	ctrl := makeController(&mockPaymentRepo{}, orderRepo, &mockMPClient{}, "")
	handler, token := withClaims(t, http.HandlerFunc(ctrl.Create), "user", "user-1")

	req := httptest.NewRequest(http.MethodPost, "/payments", jsonBody(t, map[string]any{
		"order_ids": []string{"nonexistent"},
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestCreate_PaymentAlreadyExists(t *testing.T) {
	order := makePendingOrder("order-1", "user-1")
	existing := makeTestPayment("pay-old", "order-1", "user-1")
	orderRepo := &mockOrderRepo{
		FindByIDFn: func(_ context.Context, _ string) (*orderDomain.Order, error) {
			return order, nil
		},
	}
	payRepo := &mockPaymentRepo{
		FindByOrderIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return existing, nil
		},
	}
	ctrl := makeController(payRepo, orderRepo, &mockMPClient{}, "")
	handler, token := withClaims(t, http.HandlerFunc(ctrl.Create), "user", "user-1")

	req := httptest.NewRequest(http.MethodPost, "/payments", jsonBody(t, map[string]any{
		"order_ids": []string{"order-1"},
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

// --- GetByID ---

func TestGetByID_HappyPath(t *testing.T) {
	p := makeTestPayment("pay-1", "order-1", "user-1")
	payRepo := &mockPaymentRepo{
		FindByIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return p, nil
		},
	}
	ctrl := makeController(payRepo, &mockOrderRepo{}, &mockMPClient{}, "")
	handler, token := withClaims(t, http.HandlerFunc(ctrl.GetByID), "user", "user-1")

	req := httptest.NewRequest(http.MethodGet, "/payments/pay-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "pay-1")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "pay-1", resp["id"])
}

func TestGetByID_NotFound(t *testing.T) {
	payRepo := &mockPaymentRepo{
		FindByIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return nil, apiDomain.ErrPaymentNotFound
		},
	}
	ctrl := makeController(payRepo, &mockOrderRepo{}, &mockMPClient{}, "")
	handler, token := withClaims(t, http.HandlerFunc(ctrl.GetByID), "user", "user-1")

	req := httptest.NewRequest(http.MethodGet, "/payments/nope", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "nope")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGetByID_WrongUser_Forbidden(t *testing.T) {
	p := makeTestPayment("pay-1", "order-1", "owner-user")
	payRepo := &mockPaymentRepo{
		FindByIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return p, nil
		},
	}
	ctrl := makeController(payRepo, &mockOrderRepo{}, &mockMPClient{}, "")
	handler, token := withClaims(t, http.HandlerFunc(ctrl.GetByID), "user", "other-user")

	req := httptest.NewRequest(http.MethodGet, "/payments/pay-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "pay-1")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestGetByID_NoAuth(t *testing.T) {
	ctrl := makeController(&mockPaymentRepo{}, &mockOrderRepo{}, &mockMPClient{}, "")

	req := httptest.NewRequest(http.MethodGet, "/payments/pay-1", nil)
	req.SetPathValue("id", "pay-1")
	rr := httptest.NewRecorder()

	ctrl.GetByID(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// --- GetByOrderID ---

func TestGetByOrderID_HappyPath(t *testing.T) {
	p := makeTestPayment("pay-1", "order-1", "user-1")
	payRepo := &mockPaymentRepo{
		FindByOrderIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return p, nil
		},
	}
	ctrl := makeController(payRepo, &mockOrderRepo{}, &mockMPClient{}, "")
	handler, token := withClaims(t, http.HandlerFunc(ctrl.GetByOrderID), "user", "user-1")

	req := httptest.NewRequest(http.MethodGet, "/orders/order-1/payment", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("order_id", "order-1")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "pay-1", resp["id"])
	assert.Equal(t, "order-1", resp["order_id"])
}

func TestGetByOrderID_WrongUser_Forbidden(t *testing.T) {
	p := makeTestPayment("pay-1", "order-1", "owner-user")
	payRepo := &mockPaymentRepo{
		FindByOrderIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return p, nil
		},
	}
	ctrl := makeController(payRepo, &mockOrderRepo{}, &mockMPClient{}, "")
	handler, token := withClaims(t, http.HandlerFunc(ctrl.GetByOrderID), "user", "other-user")

	req := httptest.NewRequest(http.MethodGet, "/orders/order-1/payment", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("order_id", "order-1")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// --- HandleWebhook ---

func buildWebhookSignature(secret, paymentID, requestID, ts string) string {
	manifest := fmt.Sprintf("id:%s;request-id:%s;ts:%s;", paymentID, requestID, ts)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(manifest))
	return fmt.Sprintf("ts=%s,v1=%s", ts, hex.EncodeToString(mac.Sum(nil)))
}

func TestHandleWebhook_ValidSignature_Returns200(t *testing.T) {
	secret := "webhook-secret"
	ts := "1700000000"
	requestID := "req-abc"
	mpPaymentIDStr := "42"

	payRepo := &mockPaymentRepo{
		FindByOrderIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return nil, apiDomain.ErrPaymentNotFound
		},
	}
	mpClient := &mockMPClient{
		GetPaymentFn: func(_ context.Context, _ int64) (*paymentDomain.MPPaymentResponse, error) {
			return &paymentDomain.MPPaymentResponse{
				ID:                42,
				Status:            paymentDomain.StatusApproved,
				ExternalReference: "order-1",
			}, nil
		},
	}
	ctrl := makeController(payRepo, &mockOrderRepo{}, mpClient, secret)

	body := map[string]any{
		"type":   "payment",
		"action": "payment.updated",
		"data":   map[string]string{"id": mpPaymentIDStr},
	}
	req := httptest.NewRequest(http.MethodPost, "/webhooks/mercadopago", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-signature", buildWebhookSignature(secret, mpPaymentIDStr, requestID, ts))
	req.Header.Set("x-request-id", requestID)
	rr := httptest.NewRecorder()

	ctrl.HandleWebhook(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandleWebhook_InvalidSignature_Returns200(t *testing.T) {
	secret := "webhook-secret"
	ctrl := makeController(&mockPaymentRepo{}, &mockOrderRepo{}, &mockMPClient{}, secret)

	body := map[string]any{
		"type":   "payment",
		"action": "payment.updated",
		"data":   map[string]string{"id": "42"},
	}
	req := httptest.NewRequest(http.MethodPost, "/webhooks/mercadopago", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-signature", "ts=123,v1=badsignature")
	req.Header.Set("x-request-id", "req-123")
	rr := httptest.NewRecorder()

	ctrl.HandleWebhook(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandleWebhook_NonPaymentType_Returns200(t *testing.T) {
	ctrl := makeController(&mockPaymentRepo{}, &mockOrderRepo{}, &mockMPClient{}, "")

	body := map[string]any{
		"type":   "merchant_order",
		"action": "created",
		"data":   map[string]string{"id": "99"},
	}
	req := httptest.NewRequest(http.MethodPost, "/webhooks/mercadopago", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	ctrl.HandleWebhook(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandleWebhook_InvalidJSON_Returns200(t *testing.T) {
	ctrl := makeController(&mockPaymentRepo{}, &mockOrderRepo{}, &mockMPClient{}, "")

	req := httptest.NewRequest(http.MethodPost, "/webhooks/mercadopago", bytes.NewReader([]byte("bad json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	ctrl.HandleWebhook(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandleWebhook_NoSecret_SkipsValidation(t *testing.T) {
	payRepo := &mockPaymentRepo{
		FindByOrderIDFn: func(_ context.Context, _ string) (*paymentDomain.Payment, error) {
			return nil, apiDomain.ErrPaymentNotFound
		},
	}
	mpClient := &mockMPClient{
		GetPaymentFn: func(_ context.Context, _ int64) (*paymentDomain.MPPaymentResponse, error) {
			return &paymentDomain.MPPaymentResponse{
				ID:                7,
				Status:            paymentDomain.StatusPending,
				ExternalReference: "order-7",
			}, nil
		},
	}
	ctrl := makeController(payRepo, &mockOrderRepo{}, mpClient, "") // no secret

	body := map[string]any{
		"type":   "payment",
		"action": "payment.updated",
		"data":   map[string]string{"id": "7"},
	}
	req := httptest.NewRequest(http.MethodPost, "/webhooks/mercadopago", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	ctrl.HandleWebhook(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
