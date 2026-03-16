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

	"github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	authDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/auth/domain"
	authUtils "github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/utils"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/application"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure/controller"
	storeDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
)

// --- Mocks ---

type mockStoreRepo struct {
	SaveFn     func(ctx context.Context, store *storeDomain.Store) error
	FindByIDFn func(ctx context.Context, id string) (*storeDomain.Store, error)
	FindAllFn  func(ctx context.Context, limit, offset int) ([]*storeDomain.Store, error)
	UpdateFn   func(ctx context.Context, store *storeDomain.Store) error
	DeleteFn   func(ctx context.Context, id string) error
}

func (m *mockStoreRepo) Save(ctx context.Context, store *storeDomain.Store) error {
	return m.SaveFn(ctx, store)
}
func (m *mockStoreRepo) FindByID(ctx context.Context, id string) (*storeDomain.Store, error) {
	return m.FindByIDFn(ctx, id)
}
func (m *mockStoreRepo) FindAll(ctx context.Context, limit, offset int) ([]*storeDomain.Store, error) {
	return m.FindAllFn(ctx, limit, offset)
}
func (m *mockStoreRepo) Update(ctx context.Context, store *storeDomain.Store) error {
	return m.UpdateFn(ctx, store)
}
func (m *mockStoreRepo) Delete(ctx context.Context, id string) error { return m.DeleteFn(ctx, id) }

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

func makeStoreController(repo *mockStoreRepo) *controller.StoreController {
	svc := application.NewService(repo, &mockIDGen{id: "generated-id"})
	return controller.NewStoreController(svc)
}

func jsonBody(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewReader(b)
}

func newTestStore(id string) *storeDomain.Store {
	s, _ := storeDomain.NewStore(id, "Test Store", "A test store description", "123 Test St", "12345678")
	return s
}

// --- GetByID ---

func TestGetByID_Found(t *testing.T) {
	store := newTestStore("store-123")
	repo := &mockStoreRepo{
		FindByIDFn: func(ctx context.Context, id string) (*storeDomain.Store, error) {
			return store, nil
		},
	}
	c := makeStoreController(repo)

	req := httptest.NewRequest(http.MethodGet, "/stores/store-123", nil)
	req.SetPathValue("id", "store-123")
	rr := httptest.NewRecorder()

	c.GetByID(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "store-123", resp["id"])
	assert.Equal(t, "Test Store", resp["name"])
}

func TestGetByID_NotFound(t *testing.T) {
	repo := &mockStoreRepo{
		FindByIDFn: func(ctx context.Context, id string) (*storeDomain.Store, error) {
			return nil, storeDomain.ErrStoreNotFound
		},
	}
	c := makeStoreController(repo)

	req := httptest.NewRequest(http.MethodGet, "/stores/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	rr := httptest.NewRecorder()

	c.GetByID(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// --- GetAll ---

func TestGetAll(t *testing.T) {
	stores := []*storeDomain.Store{
		newTestStore("store-1"),
		newTestStore("store-2"),
	}
	repo := &mockStoreRepo{
		FindAllFn: func(ctx context.Context, limit, offset int) ([]*storeDomain.Store, error) {
			return stores, nil
		},
	}
	c := makeStoreController(repo)

	req := httptest.NewRequest(http.MethodGet, "/stores", nil)
	rr := httptest.NewRecorder()

	c.GetAll(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp []map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Len(t, resp, 2)
}

// --- Update ---

func TestUpdate_OwnStore(t *testing.T) {
	store := newTestStore("store-123")
	repo := &mockStoreRepo{
		FindByIDFn: func(ctx context.Context, id string) (*storeDomain.Store, error) {
			return store, nil
		},
		UpdateFn: func(ctx context.Context, s *storeDomain.Store) error { return nil },
	}
	c := makeStoreController(repo)
	handler, token := withClaims(t, http.HandlerFunc(c.Update), "store", "store-123")

	req := httptest.NewRequest(http.MethodPut, "/stores/store-123", jsonBody(t, map[string]string{
		"name":         "Updated Store",
		"description":  "An updated store description",
		"address":      "456 New St",
		"phone_number": "987654321",
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "store-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "Updated Store", resp["name"])
}

func TestUpdate_NoAuth(t *testing.T) {
	c := makeStoreController(&mockStoreRepo{})

	req := httptest.NewRequest(http.MethodPut, "/stores/store-123", jsonBody(t, map[string]string{
		"name": "Updated Store",
	}))
	req.SetPathValue("id", "store-123")
	rr := httptest.NewRecorder()

	c.Update(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestUpdate_WrongAccountType(t *testing.T) {
	c := makeStoreController(&mockStoreRepo{})
	handler, token := withClaims(t, http.HandlerFunc(c.Update), "user", "store-123")

	req := httptest.NewRequest(http.MethodPut, "/stores/store-123", jsonBody(t, map[string]string{
		"name": "Updated Store",
	}))
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "store-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestUpdate_WrongStoreID(t *testing.T) {
	c := makeStoreController(&mockStoreRepo{})
	handler, token := withClaims(t, http.HandlerFunc(c.Update), "store", "other-store-456")

	req := httptest.NewRequest(http.MethodPut, "/stores/store-123", jsonBody(t, map[string]string{
		"name": "Updated Store",
	}))
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "store-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// --- Delete ---

func TestDelete_OwnStore(t *testing.T) {
	repo := &mockStoreRepo{
		DeleteFn: func(ctx context.Context, id string) error { return nil },
	}
	c := makeStoreController(repo)
	handler, token := withClaims(t, http.HandlerFunc(c.Delete), "store", "store-123")

	req := httptest.NewRequest(http.MethodDelete, "/stores/store-123", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "store-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestDelete_NoAuth(t *testing.T) {
	c := makeStoreController(&mockStoreRepo{})

	req := httptest.NewRequest(http.MethodDelete, "/stores/store-123", nil)
	req.SetPathValue("id", "store-123")
	rr := httptest.NewRecorder()

	c.Delete(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestDelete_WrongStore(t *testing.T) {
	c := makeStoreController(&mockStoreRepo{})
	handler, token := withClaims(t, http.HandlerFunc(c.Delete), "store", "other-store-456")

	req := httptest.NewRequest(http.MethodDelete, "/stores/store-123", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "store-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestDelete_WrongAccountType(t *testing.T) {
	c := makeStoreController(&mockStoreRepo{})
	handler, token := withClaims(t, http.HandlerFunc(c.Delete), "user", "store-123")

	req := httptest.NewRequest(http.MethodDelete, "/stores/store-123", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "store-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}
