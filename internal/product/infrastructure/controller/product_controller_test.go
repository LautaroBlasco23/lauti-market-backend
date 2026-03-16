package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
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
	imageDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/image/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/product/application"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/product/infrastructure/controller"
	productDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/product/domain"
	storeDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
)

// --- Mocks ---

type mockProductRepo struct {
	SaveFn          func(ctx context.Context, product *productDomain.Product) error
	FindByIDFn      func(ctx context.Context, id string) (*productDomain.Product, error)
	FindAllFn       func(ctx context.Context, limit, offset int, category *string) ([]*productDomain.Product, error)
	FindByStoreIDFn func(ctx context.Context, storeID string, limit, offset int) ([]*productDomain.Product, error)
	UpdateFn        func(ctx context.Context, product *productDomain.Product) error
	DeleteFn        func(ctx context.Context, id string) error
}

func (m *mockProductRepo) Save(ctx context.Context, p *productDomain.Product) error {
	return m.SaveFn(ctx, p)
}
func (m *mockProductRepo) FindByID(ctx context.Context, id string) (*productDomain.Product, error) {
	return m.FindByIDFn(ctx, id)
}
func (m *mockProductRepo) FindAll(ctx context.Context, limit, offset int, category *string) ([]*productDomain.Product, error) {
	return m.FindAllFn(ctx, limit, offset, category)
}
func (m *mockProductRepo) FindByStoreID(ctx context.Context, storeID string, limit, offset int) ([]*productDomain.Product, error) {
	return m.FindByStoreIDFn(ctx, storeID, limit, offset)
}
func (m *mockProductRepo) Update(ctx context.Context, p *productDomain.Product) error {
	return m.UpdateFn(ctx, p)
}
func (m *mockProductRepo) Delete(ctx context.Context, id string) error { return m.DeleteFn(ctx, id) }

type mockStoreRepo struct {
	FindByIDFn func(ctx context.Context, id string) (*storeDomain.Store, error)
}

func (m *mockStoreRepo) Save(ctx context.Context, store *storeDomain.Store) error { return nil }
func (m *mockStoreRepo) FindByID(ctx context.Context, id string) (*storeDomain.Store, error) {
	return m.FindByIDFn(ctx, id)
}
func (m *mockStoreRepo) FindAll(ctx context.Context, limit, offset int) ([]*storeDomain.Store, error) {
	return nil, nil
}
func (m *mockStoreRepo) Update(ctx context.Context, store *storeDomain.Store) error { return nil }
func (m *mockStoreRepo) Delete(ctx context.Context, id string) error                { return nil }

type mockImageClient struct {
	UploadFn func(ctx context.Context, input imageDomain.UploadImageInput) (*imageDomain.UploadImageResult, error)
}

func (m *mockImageClient) UploadImage(ctx context.Context, input imageDomain.UploadImageInput) (*imageDomain.UploadImageResult, error) {
	return m.UploadFn(ctx, input)
}

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

func makeProductController(productRepo *mockProductRepo, storeRepo *mockStoreRepo, imageClient imageDomain.ImageClient) *controller.ProductController {
	svc := application.NewService(productRepo, storeRepo, &mockIDGen{id: "prod-id"}, imageClient)
	return controller.NewProductController(svc)
}

func jsonBody(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewReader(b)
}

func newTestProduct(id, storeID string) *productDomain.Product {
	p, _ := productDomain.NewProduct(id, storeID, "Test Product", "A test product description", "electronics", 10, 99.99, nil)
	return p
}

func newTestStore(id string) *storeDomain.Store {
	s, _ := storeDomain.NewStore(id, "Test Store", "A test store description", "123 Test St", "12345678")
	return s
}

// multipartForm creates a multipart form body for product creation.
func multipartForm(t *testing.T, fields map[string]string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		require.NoError(t, w.WriteField(k, v))
	}
	require.NoError(t, w.Close())
	return &buf, w.FormDataContentType()
}

// --- GetByID ---

func TestGetByID_Found(t *testing.T) {
	product := newTestProduct("prod-123", "store-123")
	repo := &mockProductRepo{
		FindByIDFn: func(ctx context.Context, id string) (*productDomain.Product, error) {
			return product, nil
		},
	}
	c := makeProductController(repo, &mockStoreRepo{}, &mockImageClient{})

	req := httptest.NewRequest(http.MethodGet, "/stores/store-123/products/prod-123", nil)
	req.SetPathValue("id", "prod-123")
	req.SetPathValue("store_id", "store-123")
	rr := httptest.NewRecorder()

	c.GetByID(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "prod-123", resp["id"])
	assert.Equal(t, "Test Product", resp["name"])
}

func TestGetByID_NotFound(t *testing.T) {
	repo := &mockProductRepo{
		FindByIDFn: func(ctx context.Context, id string) (*productDomain.Product, error) {
			return nil, apiDomain.ErrProductNotFound
		},
	}
	c := makeProductController(repo, &mockStoreRepo{}, &mockImageClient{})

	req := httptest.NewRequest(http.MethodGet, "/stores/store-123/products/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	rr := httptest.NewRecorder()

	c.GetByID(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// --- GetAll ---

func TestGetAll(t *testing.T) {
	products := []*productDomain.Product{
		newTestProduct("prod-1", "store-1"),
		newTestProduct("prod-2", "store-1"),
	}
	repo := &mockProductRepo{
		FindAllFn: func(ctx context.Context, limit, offset int, category *string) ([]*productDomain.Product, error) {
			return products, nil
		},
	}
	c := makeProductController(repo, &mockStoreRepo{}, &mockImageClient{})

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rr := httptest.NewRecorder()

	c.GetAll(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	products2, ok := resp["products"].([]any)
	require.True(t, ok)
	assert.Len(t, products2, 2)
}

// --- GetByStoreID ---

func TestGetByStoreID(t *testing.T) {
	store := newTestStore("store-123")
	products := []*productDomain.Product{
		newTestProduct("prod-1", "store-123"),
	}
	storeRepo := &mockStoreRepo{
		FindByIDFn: func(ctx context.Context, id string) (*storeDomain.Store, error) {
			return store, nil
		},
	}
	productRepo := &mockProductRepo{
		FindByStoreIDFn: func(ctx context.Context, storeID string, limit, offset int) ([]*productDomain.Product, error) {
			return products, nil
		},
	}
	c := makeProductController(productRepo, storeRepo, &mockImageClient{})

	req := httptest.NewRequest(http.MethodGet, "/stores/store-123/products", nil)
	req.SetPathValue("store_id", "store-123")
	rr := httptest.NewRecorder()

	c.GetByStoreID(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp []any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Len(t, resp, 1)
}

// --- Create ---

func TestCreate_StoreOwner(t *testing.T) {
	store := newTestStore("store-123")
	storeRepo := &mockStoreRepo{
		FindByIDFn: func(ctx context.Context, id string) (*storeDomain.Store, error) {
			return store, nil
		},
	}
	productRepo := &mockProductRepo{
		SaveFn: func(ctx context.Context, p *productDomain.Product) error { return nil },
	}
	c := makeProductController(productRepo, storeRepo, &mockImageClient{})
	handler, token := withClaims(t, http.HandlerFunc(c.Create), "store", "store-123")

	body, contentType := multipartForm(t, map[string]string{
		"name":        "New Product",
		"description": "A great new product for testing",
		"category":    "electronics",
		"stock":       "5",
		"price":       "49.99",
	})
	req := httptest.NewRequest(http.MethodPost, "/stores/store-123/products", body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("store_id", "store-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "New Product", resp["name"])
}

func TestCreate_NoAuth(t *testing.T) {
	c := makeProductController(&mockProductRepo{}, &mockStoreRepo{}, &mockImageClient{})

	body, contentType := multipartForm(t, map[string]string{
		"name": "New Product", "description": "desc", "category": "cat", "stock": "5", "price": "9.99",
	})
	req := httptest.NewRequest(http.MethodPost, "/stores/store-123/products", body)
	req.Header.Set("Content-Type", contentType)
	req.SetPathValue("store_id", "store-123")
	rr := httptest.NewRecorder()

	c.Create(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestCreate_WrongAccountType(t *testing.T) {
	c := makeProductController(&mockProductRepo{}, &mockStoreRepo{}, &mockImageClient{})
	handler, token := withClaims(t, http.HandlerFunc(c.Create), "user", "store-123")

	body, contentType := multipartForm(t, map[string]string{
		"name": "New Product", "description": "desc", "category": "cat", "stock": "5", "price": "9.99",
	})
	req := httptest.NewRequest(http.MethodPost, "/stores/store-123/products", body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("store_id", "store-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestCreate_WrongStoreID(t *testing.T) {
	c := makeProductController(&mockProductRepo{}, &mockStoreRepo{}, &mockImageClient{})
	handler, token := withClaims(t, http.HandlerFunc(c.Create), "store", "other-store-456")

	body, contentType := multipartForm(t, map[string]string{
		"name": "New Product", "description": "desc", "category": "cat", "stock": "5", "price": "9.99",
	})
	req := httptest.NewRequest(http.MethodPost, "/stores/store-123/products", body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("store_id", "store-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// --- Update ---

func TestUpdate_StoreOwner(t *testing.T) {
	product := newTestProduct("prod-123", "store-123")
	productRepo := &mockProductRepo{
		FindByIDFn: func(ctx context.Context, id string) (*productDomain.Product, error) {
			return product, nil
		},
		UpdateFn: func(ctx context.Context, p *productDomain.Product) error { return nil },
	}
	c := makeProductController(productRepo, &mockStoreRepo{}, &mockImageClient{})
	handler, token := withClaims(t, http.HandlerFunc(c.Update), "store", "store-123")

	req := httptest.NewRequest(http.MethodPut, "/stores/store-123/products/prod-123", jsonBody(t, map[string]any{
		"name":        "Updated Product",
		"description": "An updated product description",
		"category":    "electronics",
		"stock":       20,
		"price":       129.99,
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("store_id", "store-123")
	req.SetPathValue("id", "prod-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestUpdate_WrongStore(t *testing.T) {
	c := makeProductController(&mockProductRepo{}, &mockStoreRepo{}, &mockImageClient{})
	handler, token := withClaims(t, http.HandlerFunc(c.Update), "store", "other-store")

	req := httptest.NewRequest(http.MethodPut, "/stores/store-123/products/prod-123", jsonBody(t, map[string]any{
		"name": "Updated", "description": "desc", "category": "cat", "stock": 1, "price": 9.99,
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("store_id", "store-123")
	req.SetPathValue("id", "prod-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// --- Delete ---

func TestDelete_StoreOwner(t *testing.T) {
	productRepo := &mockProductRepo{
		DeleteFn: func(ctx context.Context, id string) error { return nil },
	}
	c := makeProductController(productRepo, &mockStoreRepo{}, &mockImageClient{})
	handler, token := withClaims(t, http.HandlerFunc(c.Delete), "store", "store-123")

	req := httptest.NewRequest(http.MethodDelete, "/stores/store-123/products/prod-123", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("store_id", "store-123")
	req.SetPathValue("id", "prod-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestDelete_NoAuth(t *testing.T) {
	c := makeProductController(&mockProductRepo{}, &mockStoreRepo{}, &mockImageClient{})

	req := httptest.NewRequest(http.MethodDelete, "/stores/store-123/products/prod-123", nil)
	req.SetPathValue("store_id", "store-123")
	req.SetPathValue("id", "prod-123")
	rr := httptest.NewRecorder()

	c.Delete(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestDelete_WrongStore(t *testing.T) {
	c := makeProductController(&mockProductRepo{}, &mockStoreRepo{}, &mockImageClient{})
	handler, token := withClaims(t, http.HandlerFunc(c.Delete), "store", "other-store")

	req := httptest.NewRequest(http.MethodDelete, "/stores/store-123/products/prod-123", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("store_id", "store-123")
	req.SetPathValue("id", "prod-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

