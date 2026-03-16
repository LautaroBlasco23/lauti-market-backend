package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	authDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/auth/domain"
	authUtils "github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/utils"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/order/application"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/order/infrastructure/controller"
	orderDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/order/domain"
	productDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/product/domain"
)

// --- Mocks ---

type mockOrderRepo struct {
	SaveFn          func(ctx context.Context, order *orderDomain.Order) error
	FindByIDFn      func(ctx context.Context, id string) (*orderDomain.Order, error)
	FindByUserIDFn  func(ctx context.Context, userID string, limit, offset int) ([]*orderDomain.Order, error)
	FindByStoreIDFn func(ctx context.Context, storeID string, limit, offset int) ([]*orderDomain.Order, error)
	UpdateStatusFn  func(ctx context.Context, id string, status orderDomain.OrderStatus) error
}

func (m *mockOrderRepo) Save(ctx context.Context, order *orderDomain.Order) error {
	return m.SaveFn(ctx, order)
}
func (m *mockOrderRepo) FindByID(ctx context.Context, id string) (*orderDomain.Order, error) {
	return m.FindByIDFn(ctx, id)
}
func (m *mockOrderRepo) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*orderDomain.Order, error) {
	return m.FindByUserIDFn(ctx, userID, limit, offset)
}
func (m *mockOrderRepo) FindByStoreID(ctx context.Context, storeID string, limit, offset int) ([]*orderDomain.Order, error) {
	return m.FindByStoreIDFn(ctx, storeID, limit, offset)
}
func (m *mockOrderRepo) UpdateStatus(ctx context.Context, id string, status orderDomain.OrderStatus) error {
	return m.UpdateStatusFn(ctx, id, status)
}

type mockProductRepo struct {
	FindByIDFn func(ctx context.Context, id string) (*productDomain.Product, error)
	UpdateFn   func(ctx context.Context, product *productDomain.Product) error
}

func (m *mockProductRepo) Save(ctx context.Context, p *productDomain.Product) error { return nil }
func (m *mockProductRepo) FindByID(ctx context.Context, id string) (*productDomain.Product, error) {
	return m.FindByIDFn(ctx, id)
}
func (m *mockProductRepo) FindAll(ctx context.Context, limit, offset int, category *string) ([]*productDomain.Product, error) {
	return nil, nil
}
func (m *mockProductRepo) FindByStoreID(ctx context.Context, storeID string, limit, offset int) ([]*productDomain.Product, error) {
	return nil, nil
}
func (m *mockProductRepo) Update(ctx context.Context, p *productDomain.Product) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, p)
	}
	return nil
}
func (m *mockProductRepo) Delete(ctx context.Context, id string) error { return nil }

type mockIDGen struct{ id string }

func (m *mockIDGen) Generate() string { return m.id }

func testJWTGen() *authUtils.JWTGenerator {
	return authUtils.NewJWTGenerator("test-secret", time.Hour)
}

func withClaims(t *testing.T, handler http.Handler, accountType, accountID string) (http.Handler, string) {
	t.Helper()
	jwtGen := testJWTGen()
	token, err := jwtGen.Generate("auth-1", authDomain.AccountType(accountType), accountID)
	require.NoError(t, err)
	mw := infrastructure.NewAuthMiddleware(jwtGen)
	return mw.Wrap(handler), token
}

type seqIDGen struct {
	ids []string
	idx int
}

func (g *seqIDGen) Generate() string {
	id := g.ids[g.idx]
	g.idx++
	return id
}

func makeOrderController(orderRepo *mockOrderRepo, productRepo *mockProductRepo, idGen apiDomain.IDGenerator) *controller.OrderController {
	svc := application.NewOrderService(orderRepo, productRepo, idGen)
	return controller.NewOrderController(svc)
}

func jsonBody(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewReader(b)
}

func newTestOrder(id, userID, storeID string) *orderDomain.Order {
	item, _ := orderDomain.NewOrderItem("item-1", id, "prod-1", 2, 50.0)
	order, _ := orderDomain.NewOrder(id, userID, storeID, []*orderDomain.OrderItem{item}, 100.0)
	return order
}

// --- Create ---

func TestCreate_UserAccount(t *testing.T) {
	product, _ := productDomain.NewProduct("prod-1", "store-123", "Test", "A test product desc", "cat", 10, 50.0, nil)
	productRepo := &mockProductRepo{
		FindByIDFn: func(ctx context.Context, id string) (*productDomain.Product, error) {
			return product, nil
		},
	}
	orderRepo := &mockOrderRepo{
		SaveFn: func(ctx context.Context, order *orderDomain.Order) error { return nil },
	}
	c := makeOrderController(orderRepo, productRepo, &seqIDGen{ids: []string{"order-1", "item-1"}})
	handler, token := withClaims(t, http.HandlerFunc(c.Create), "user", "user-123")

	req := httptest.NewRequest(http.MethodPost, "/orders", jsonBody(t, map[string]any{
		"store_id": "store-123",
		"items":    []map[string]any{{"product_id": "prod-1", "quantity": 2}},
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "user-123", resp["user_id"])
	assert.Equal(t, "pending", resp["status"])
}

func TestCreate_StoreAccountForbidden(t *testing.T) {
	c := makeOrderController(&mockOrderRepo{}, &mockProductRepo{}, &mockIDGen{})
	handler, token := withClaims(t, http.HandlerFunc(c.Create), "store", "store-123")

	req := httptest.NewRequest(http.MethodPost, "/orders", jsonBody(t, map[string]any{
		"store_id": "store-123",
		"items":    []map[string]any{{"product_id": "prod-1", "quantity": 1}},
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestCreate_NoAuth(t *testing.T) {
	c := makeOrderController(&mockOrderRepo{}, &mockProductRepo{}, &mockIDGen{})

	req := httptest.NewRequest(http.MethodPost, "/orders", jsonBody(t, map[string]any{
		"store_id": "store-123",
		"items":    []map[string]any{{"product_id": "prod-1", "quantity": 1}},
	}))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c.Create(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// --- GetByID ---

func TestGetByID_UserOwner(t *testing.T) {
	order := newTestOrder("order-123", "user-123", "store-456")
	orderRepo := &mockOrderRepo{
		FindByIDFn: func(ctx context.Context, id string) (*orderDomain.Order, error) {
			return order, nil
		},
	}
	c := makeOrderController(orderRepo, &mockProductRepo{}, &mockIDGen{})
	handler, token := withClaims(t, http.HandlerFunc(c.GetByID), "user", "user-123")

	req := httptest.NewRequest(http.MethodGet, "/orders/order-123", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "order-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "order-123", resp["id"])
}

func TestGetByID_StoreOwner(t *testing.T) {
	order := newTestOrder("order-123", "user-123", "store-456")
	orderRepo := &mockOrderRepo{
		FindByIDFn: func(ctx context.Context, id string) (*orderDomain.Order, error) {
			return order, nil
		},
	}
	c := makeOrderController(orderRepo, &mockProductRepo{}, &mockIDGen{})
	handler, token := withClaims(t, http.HandlerFunc(c.GetByID), "store", "store-456")

	req := httptest.NewRequest(http.MethodGet, "/orders/order-123", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "order-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetByID_NotOwner(t *testing.T) {
	order := newTestOrder("order-123", "user-123", "store-456")
	orderRepo := &mockOrderRepo{
		FindByIDFn: func(ctx context.Context, id string) (*orderDomain.Order, error) {
			return order, nil
		},
	}
	c := makeOrderController(orderRepo, &mockProductRepo{}, &mockIDGen{})
	handler, token := withClaims(t, http.HandlerFunc(c.GetByID), "user", "other-user-789")

	req := httptest.NewRequest(http.MethodGet, "/orders/order-123", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "order-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestGetByID_NotFound(t *testing.T) {
	orderRepo := &mockOrderRepo{
		FindByIDFn: func(ctx context.Context, id string) (*orderDomain.Order, error) {
			return nil, apiDomain.ErrOrderNotFound
		},
	}
	c := makeOrderController(orderRepo, &mockProductRepo{}, &mockIDGen{})
	handler, token := withClaims(t, http.HandlerFunc(c.GetByID), "user", "user-123")

	req := httptest.NewRequest(http.MethodGet, "/orders/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "nonexistent")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// --- GetByUserID ---

func TestGetByUserID_OwnUser(t *testing.T) {
	orders := []*orderDomain.Order{newTestOrder("order-1", "user-123", "store-456")}
	orderRepo := &mockOrderRepo{
		FindByUserIDFn: func(ctx context.Context, userID string, limit, offset int) ([]*orderDomain.Order, error) {
			return orders, nil
		},
	}
	c := makeOrderController(orderRepo, &mockProductRepo{}, &mockIDGen{})
	handler, token := withClaims(t, http.HandlerFunc(c.GetByUserID), "user", "user-123")

	req := httptest.NewRequest(http.MethodGet, "/users/user-123/orders", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("user_id", "user-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp []any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Len(t, resp, 1)
}

func TestGetByUserID_WrongUser(t *testing.T) {
	c := makeOrderController(&mockOrderRepo{}, &mockProductRepo{}, &mockIDGen{})
	handler, token := withClaims(t, http.HandlerFunc(c.GetByUserID), "user", "other-user-456")

	req := httptest.NewRequest(http.MethodGet, "/users/user-123/orders", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("user_id", "user-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestGetByUserID_StoreAccountForbidden(t *testing.T) {
	c := makeOrderController(&mockOrderRepo{}, &mockProductRepo{}, &mockIDGen{})
	handler, token := withClaims(t, http.HandlerFunc(c.GetByUserID), "store", "user-123")

	req := httptest.NewRequest(http.MethodGet, "/users/user-123/orders", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("user_id", "user-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// --- GetByStoreID ---

func TestGetByStoreID_OwnStore(t *testing.T) {
	orders := []*orderDomain.Order{newTestOrder("order-1", "user-123", "store-456")}
	orderRepo := &mockOrderRepo{
		FindByStoreIDFn: func(ctx context.Context, storeID string, limit, offset int) ([]*orderDomain.Order, error) {
			return orders, nil
		},
	}
	c := makeOrderController(orderRepo, &mockProductRepo{}, &mockIDGen{})
	handler, token := withClaims(t, http.HandlerFunc(c.GetByStoreID), "store", "store-456")

	req := httptest.NewRequest(http.MethodGet, "/stores/store-456/orders", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("store_id", "store-456")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetByStoreID_WrongStore(t *testing.T) {
	c := makeOrderController(&mockOrderRepo{}, &mockProductRepo{}, &mockIDGen{})
	handler, token := withClaims(t, http.HandlerFunc(c.GetByStoreID), "store", "other-store")

	req := httptest.NewRequest(http.MethodGet, "/stores/store-456/orders", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("store_id", "store-456")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// --- UpdateStatus ---

func TestUpdateStatus_ConfirmByStore(t *testing.T) {
	order := newTestOrder("order-123", "user-123", "store-456")
	orderRepo := &mockOrderRepo{
		FindByIDFn: func(ctx context.Context, id string) (*orderDomain.Order, error) {
			return order, nil
		},
		UpdateStatusFn: func(ctx context.Context, id string, status orderDomain.OrderStatus) error {
			return nil
		},
	}
	c := makeOrderController(orderRepo, &mockProductRepo{}, &mockIDGen{})
	handler, token := withClaims(t, http.HandlerFunc(c.UpdateStatus), "store", "store-456")

	req := httptest.NewRequest(http.MethodPut, "/orders/order-123/status", jsonBody(t, map[string]string{
		"status": "confirmed",
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "order-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "confirmed", resp["status"])
}

func TestUpdateStatus_CancelByUser(t *testing.T) {
	order := newTestOrder("order-123", "user-123", "store-456")
	product, _ := productDomain.NewProduct("prod-1", "store-456", "Test", "A test product description", "cat", 8, 50.0, nil)
	productRepo := &mockProductRepo{
		FindByIDFn: func(ctx context.Context, id string) (*productDomain.Product, error) {
			return product, nil
		},
		UpdateFn: func(ctx context.Context, p *productDomain.Product) error { return nil },
	}
	orderRepo := &mockOrderRepo{
		FindByIDFn: func(ctx context.Context, id string) (*orderDomain.Order, error) {
			return order, nil
		},
		UpdateStatusFn: func(ctx context.Context, id string, status orderDomain.OrderStatus) error {
			return nil
		},
	}
	c := makeOrderController(orderRepo, productRepo, &mockIDGen{})
	handler, token := withClaims(t, http.HandlerFunc(c.UpdateStatus), "user", "user-123")

	req := httptest.NewRequest(http.MethodPut, "/orders/order-123/status", jsonBody(t, map[string]string{
		"status": "cancelled",
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "order-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestUpdateStatus_InvalidStatus(t *testing.T) {
	c := makeOrderController(&mockOrderRepo{}, &mockProductRepo{}, &mockIDGen{})
	handler, token := withClaims(t, http.HandlerFunc(c.UpdateStatus), "store", "store-456")

	req := httptest.NewRequest(http.MethodPut, "/orders/order-123/status", jsonBody(t, map[string]string{
		"status": "flying",
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "order-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUpdateStatus_NoAuth(t *testing.T) {
	c := makeOrderController(&mockOrderRepo{}, &mockProductRepo{}, &mockIDGen{})

	req := httptest.NewRequest(http.MethodPut, "/orders/order-123/status", jsonBody(t, map[string]string{
		"status": "confirmed",
	}))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "order-123")
	rr := httptest.NewRecorder()

	c.UpdateStatus(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
