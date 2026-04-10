package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/application"
	authDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/auth/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/controller"
	storeApp "github.com/LautaroBlasco23/lauti-market-backend/internal/store/application"
	storeDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure/mercadopago"
	userApp "github.com/LautaroBlasco23/lauti-market-backend/internal/user/application"
	userDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/user/domain"
)

// --- Mocks ---

type mockIDGen struct{ id string }

func (m *mockIDGen) Generate() string { return m.id }

type mockAuthRepo struct {
	SaveFn        func(ctx context.Context, auth *authDomain.Auth) error
	FindByIDFn    func(ctx context.Context, id string) (*authDomain.Auth, error)
	FindByEmailFn func(ctx context.Context, email string) (*authDomain.Auth, error)
	DeleteFn      func(ctx context.Context, id string) error
}

func (m *mockAuthRepo) Save(ctx context.Context, auth *authDomain.Auth) error {
	return m.SaveFn(ctx, auth)
}

func (m *mockAuthRepo) FindByID(ctx context.Context, id string) (*authDomain.Auth, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *mockAuthRepo) FindByEmail(ctx context.Context, email string) (*authDomain.Auth, error) {
	return m.FindByEmailFn(ctx, email)
}
func (m *mockAuthRepo) Delete(ctx context.Context, id string) error { return m.DeleteFn(ctx, id) }

type mockUserRepo struct {
	SaveFn     func(ctx context.Context, user *userDomain.User) error
	UpdateFn   func(ctx context.Context, user *userDomain.User) error
	FindByIDFn func(ctx context.Context, id string) (*userDomain.User, error)
	DeleteFn   func(ctx context.Context, id string) error
}

func (m *mockUserRepo) Save(ctx context.Context, user *userDomain.User) error {
	return m.SaveFn(ctx, user)
}

func (m *mockUserRepo) Update(ctx context.Context, user *userDomain.User) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*userDomain.User, error) {
	return m.FindByIDFn(ctx, id)
}
func (m *mockUserRepo) Delete(ctx context.Context, id string) error { return m.DeleteFn(ctx, id) }

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
func (m *mockStoreRepo) UpdateMPConnection(ctx context.Context, storeID string, fields storeDomain.MPFields) error {
	return nil
}

type mockHasher struct {
	HashFn    func(password string) (string, error)
	CompareFn func(hashed, plain string) error
}

func (m *mockHasher) Hash(password string) (string, error) { return m.HashFn(password) }
func (m *mockHasher) Compare(hashed, plain string) error   { return m.CompareFn(hashed, plain) }

type mockTokenGen struct {
	GenerateFn func(authID string, accountType authDomain.AccountType, accountID string) (string, error)
}

func (m *mockTokenGen) Generate(authID string, accountType authDomain.AccountType, accountID string) (string, error) {
	return m.GenerateFn(authID, accountType, accountID)
}

func makeAuthController(
	authRepo *mockAuthRepo,
	userRepo *mockUserRepo,
	storeRepo *mockStoreRepo,
	hasher *mockHasher,
	tokenGen *mockTokenGen,
	idGen *mockIDGen,
) *controller.Controller {
	userSvc := userApp.NewService(userRepo, idGen)
	mpOAuth := mercadopago.NewOAuthClient("test-client-id", "test-client-secret", "http://localhost/callback")
	storeSvc := storeApp.NewService(storeRepo, idGen, mpOAuth)
	svc := application.NewService(authRepo, idGen, hasher, tokenGen, userSvc, storeSvc)
	return controller.NewController(svc)
}

func jsonBody(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewReader(b)
}

// --- Register User ---

func TestRegisterUser_HappyPath(t *testing.T) {
	authRepo := &mockAuthRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*authDomain.Auth, error) {
			return nil, apiDomain.ErrAuthNotFound
		},
		SaveFn: func(ctx context.Context, auth *authDomain.Auth) error { return nil },
	}
	userRepo := &mockUserRepo{
		SaveFn: func(ctx context.Context, user *userDomain.User) error { return nil },
	}
	hasher := &mockHasher{
		HashFn: func(password string) (string, error) { return "hashed-password", nil },
	}
	tokenGen := &mockTokenGen{
		GenerateFn: func(authID string, accountType authDomain.AccountType, accountID string) (string, error) {
			return "jwt-token", nil
		},
	}
	c := makeAuthController(authRepo, userRepo, &mockStoreRepo{}, hasher, tokenGen, &mockIDGen{id: "new-id"})

	req := httptest.NewRequest(http.MethodPost, "/auth/register/user", jsonBody(t, map[string]string{
		"email": "test@example.com", "password": "password123",
		"first_name": "John", "last_name": "Doe",
	}))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c.RegisterUser(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var resp map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "test@example.com", resp["email"])
	assert.Equal(t, "user", resp["account_type"])
	assert.NotEmpty(t, resp["auth_id"])
}

func TestRegisterUser_InvalidPayload(t *testing.T) {
	c := makeAuthController(&mockAuthRepo{}, &mockUserRepo{}, &mockStoreRepo{}, &mockHasher{}, &mockTokenGen{}, &mockIDGen{})

	req := httptest.NewRequest(http.MethodPost, "/auth/register/user", jsonBody(t, map[string]string{
		"email": "bad-email",
	}))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c.RegisterUser(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRegisterUser_EmailExists(t *testing.T) {
	existingAuth, err := authDomain.NewAuth("existing", "test@example.com", "hashed-pw1", "acc-1", authDomain.AccountTypeUser)
	require.NoError(t, err)

	authRepo := &mockAuthRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*authDomain.Auth, error) {
			return existingAuth, nil
		},
	}
	c := makeAuthController(authRepo, &mockUserRepo{}, &mockStoreRepo{}, &mockHasher{}, &mockTokenGen{}, &mockIDGen{})

	req := httptest.NewRequest(http.MethodPost, "/auth/register/user", jsonBody(t, map[string]string{
		"email": "test@example.com", "password": "password123",
		"first_name": "John", "last_name": "Doe",
	}))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c.RegisterUser(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

// --- Register Store ---

func TestRegisterStore_HappyPath(t *testing.T) {
	authRepo := &mockAuthRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*authDomain.Auth, error) {
			return nil, apiDomain.ErrAuthNotFound
		},
		SaveFn: func(ctx context.Context, auth *authDomain.Auth) error { return nil },
	}
	storeRepo := &mockStoreRepo{
		SaveFn: func(ctx context.Context, store *storeDomain.Store) error { return nil },
	}
	hasher := &mockHasher{
		HashFn: func(password string) (string, error) { return "hashed-password", nil },
	}
	tokenGen := &mockTokenGen{
		GenerateFn: func(authID string, accountType authDomain.AccountType, accountID string) (string, error) {
			return "jwt-token", nil
		},
	}
	c := makeAuthController(authRepo, &mockUserRepo{}, storeRepo, hasher, tokenGen, &mockIDGen{id: "new-id"})

	req := httptest.NewRequest(http.MethodPost, "/auth/register/store", jsonBody(t, map[string]string{
		"email":        "store@example.com",
		"password":     "password123",
		"name":         "My Store",
		"description":  "A great store with many products",
		"address":      "123 Main St",
		"phone_number": "123456789",
	}))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c.RegisterStore(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var resp map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "store@example.com", resp["email"])
	assert.Equal(t, "store", resp["account_type"])
}

func TestRegisterStore_InvalidPayload(t *testing.T) {
	c := makeAuthController(&mockAuthRepo{}, &mockUserRepo{}, &mockStoreRepo{}, &mockHasher{}, &mockTokenGen{}, &mockIDGen{})

	req := httptest.NewRequest(http.MethodPost, "/auth/register/store", jsonBody(t, map[string]string{
		"email": "store@example.com",
	}))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c.RegisterStore(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRegisterStore_EmailExists(t *testing.T) {
	existingAuth, err := authDomain.NewAuth("existing", "store@example.com", "hashed-pw1", "store-1", authDomain.AccountTypeStore)
	require.NoError(t, err)

	authRepo := &mockAuthRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*authDomain.Auth, error) {
			return existingAuth, nil
		},
	}
	c := makeAuthController(authRepo, &mockUserRepo{}, &mockStoreRepo{}, &mockHasher{}, &mockTokenGen{}, &mockIDGen{})

	req := httptest.NewRequest(http.MethodPost, "/auth/register/store", jsonBody(t, map[string]string{
		"email":        "store@example.com",
		"password":     "password123",
		"name":         "My Store",
		"description":  "A great store with many products",
		"address":      "123 Main St",
		"phone_number": "123456789",
	}))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c.RegisterStore(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

// --- Login ---

func TestLogin_HappyPath(t *testing.T) {
	existingAuth, err := authDomain.NewAuth("auth-1", "user@example.com", "hashed-pw1", "acc-1", authDomain.AccountTypeUser)
	require.NoError(t, err)

	authRepo := &mockAuthRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*authDomain.Auth, error) {
			return existingAuth, nil
		},
	}
	hasher := &mockHasher{
		CompareFn: func(hashed, plain string) error { return nil },
	}
	tokenGen := &mockTokenGen{
		GenerateFn: func(authID string, accountType authDomain.AccountType, accountID string) (string, error) {
			return "jwt-token", nil
		},
	}
	c := makeAuthController(authRepo, &mockUserRepo{}, &mockStoreRepo{}, hasher, tokenGen, &mockIDGen{})

	req := httptest.NewRequest(http.MethodPost, "/auth/login", jsonBody(t, map[string]string{
		"email":    "user@example.com",
		"password": "password123",
	}))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c.Login(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "jwt-token", resp["token"])
	assert.Equal(t, "acc-1", resp["account_id"])
}

func TestLogin_InvalidCredentials_EmailNotFound(t *testing.T) {
	authRepo := &mockAuthRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*authDomain.Auth, error) {
			return nil, apiDomain.ErrAuthNotFound
		},
	}
	c := makeAuthController(authRepo, &mockUserRepo{}, &mockStoreRepo{}, &mockHasher{}, &mockTokenGen{}, &mockIDGen{})

	req := httptest.NewRequest(http.MethodPost, "/auth/login", jsonBody(t, map[string]string{
		"email":    "nobody@example.com",
		"password": "password123",
	}))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c.Login(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestLogin_InvalidCredentials_WrongPassword(t *testing.T) {
	existingAuth, err := authDomain.NewAuth("auth-1", "user@example.com", "hashed-pw1", "acc-1", authDomain.AccountTypeUser)
	require.NoError(t, err)

	authRepo := &mockAuthRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*authDomain.Auth, error) {
			return existingAuth, nil
		},
	}
	hasher := &mockHasher{
		CompareFn: func(hashed, plain string) error { return apiDomain.ErrInvalidCredentials },
	}
	c := makeAuthController(authRepo, &mockUserRepo{}, &mockStoreRepo{}, hasher, &mockTokenGen{}, &mockIDGen{})

	req := httptest.NewRequest(http.MethodPost, "/auth/login", jsonBody(t, map[string]string{
		"email":    "user@example.com",
		"password": "wrongpassword",
	}))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c.Login(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestLogin_InvalidPayload(t *testing.T) {
	c := makeAuthController(&mockAuthRepo{}, &mockUserRepo{}, &mockStoreRepo{}, &mockHasher{}, &mockTokenGen{}, &mockIDGen{})

	req := httptest.NewRequest(http.MethodPost, "/auth/login", jsonBody(t, map[string]string{
		"email": "bad-email",
	}))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c.Login(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
