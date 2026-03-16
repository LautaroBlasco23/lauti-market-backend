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
	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/application"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/infrastructure/controller"
	userDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/user/domain"
)

// --- Mocks ---

type mockUserRepo struct {
	SaveFn     func(ctx context.Context, user *userDomain.User) error
	FindByIDFn func(ctx context.Context, id string) (*userDomain.User, error)
	DeleteFn   func(ctx context.Context, id string) error
}

func (m *mockUserRepo) Save(ctx context.Context, user *userDomain.User) error {
	return m.SaveFn(ctx, user)
}
func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*userDomain.User, error) {
	return m.FindByIDFn(ctx, id)
}
func (m *mockUserRepo) Delete(ctx context.Context, id string) error { return m.DeleteFn(ctx, id) }

type mockIDGen struct{ id string }

func (m *mockIDGen) Generate() string { return m.id }

// testJWTGen creates a JWTGenerator with a test secret for use in tests.
func testJWTGen() *authUtils.JWTGenerator {
	return authUtils.NewJWTGenerator("test-secret", time.Hour)
}

// withClaims wraps the handler with auth middleware and sets the Authorization header
// with a token for the given account type and ID.
func withClaims(t *testing.T, handler http.Handler, accountType, accountID string) (http.Handler, string) {
	t.Helper()
	jwtGen := testJWTGen()
	token, err := jwtGen.Generate("auth-1", authDomain.AccountType(accountType), accountID)
	require.NoError(t, err)
	mw := infrastructure.NewAuthMiddleware(jwtGen)
	return mw.Wrap(handler), token
}

func makeUserController(repo *mockUserRepo) *controller.UserController {
	svc := application.NewService(repo, &mockIDGen{id: "generated-id"})
	return controller.NewUserController(svc)
}

func jsonBody(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewReader(b)
}

// --- GetByID ---

func TestGetByID_Found(t *testing.T) {
	user, err := userDomain.NewUser("user-123", "John", "Doe")
	require.NoError(t, err)

	repo := &mockUserRepo{
		FindByIDFn: func(ctx context.Context, id string) (*userDomain.User, error) {
			return user, nil
		},
	}
	c := makeUserController(repo)

	req := httptest.NewRequest(http.MethodGet, "/users/user-123", nil)
	req.SetPathValue("id", "user-123")
	rr := httptest.NewRecorder()

	c.GetByID(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "user-123", resp["id"])
	assert.Equal(t, "John", resp["first_name"])
}

func TestGetByID_NotFound(t *testing.T) {
	repo := &mockUserRepo{
		FindByIDFn: func(ctx context.Context, id string) (*userDomain.User, error) {
			return nil, userDomain.ErrUserNotFound
		},
	}
	c := makeUserController(repo)

	req := httptest.NewRequest(http.MethodGet, "/users/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	rr := httptest.NewRecorder()

	c.GetByID(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// --- Update ---

func TestUpdate_OwnAccount(t *testing.T) {
	user, err := userDomain.NewUser("user-123", "John", "Doe")
	require.NoError(t, err)

	repo := &mockUserRepo{
		FindByIDFn: func(ctx context.Context, id string) (*userDomain.User, error) {
			return user, nil
		},
		SaveFn: func(ctx context.Context, u *userDomain.User) error { return nil },
	}
	c := makeUserController(repo)

	handler, token := withClaims(t, http.HandlerFunc(c.Update), "user", "user-123")

	req := httptest.NewRequest(http.MethodPut, "/users/user-123", jsonBody(t, map[string]string{
		"first_name": "Jane",
		"last_name":  "Smith",
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "user-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "Jane", resp["first_name"])
}

func TestUpdate_NoAuth(t *testing.T) {
	c := makeUserController(&mockUserRepo{})

	req := httptest.NewRequest(http.MethodPut, "/users/user-123", jsonBody(t, map[string]string{
		"first_name": "Jane", "last_name": "Smith",
	}))
	req.SetPathValue("id", "user-123")
	rr := httptest.NewRecorder()

	c.Update(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestUpdate_WrongAccountType(t *testing.T) {
	c := makeUserController(&mockUserRepo{})
	handler, token := withClaims(t, http.HandlerFunc(c.Update), "store", "user-123")

	req := httptest.NewRequest(http.MethodPut, "/users/user-123", jsonBody(t, map[string]string{
		"first_name": "Jane", "last_name": "Smith",
	}))
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "user-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestUpdate_WrongAccountID(t *testing.T) {
	c := makeUserController(&mockUserRepo{})
	handler, token := withClaims(t, http.HandlerFunc(c.Update), "user", "other-user-456")

	req := httptest.NewRequest(http.MethodPut, "/users/user-123", jsonBody(t, map[string]string{
		"first_name": "Jane", "last_name": "Smith",
	}))
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "user-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestUpdate_InvalidPayload(t *testing.T) {
	user, err := userDomain.NewUser("user-123", "John", "Doe")
	require.NoError(t, err)

	repo := &mockUserRepo{
		FindByIDFn: func(ctx context.Context, id string) (*userDomain.User, error) {
			return user, nil
		},
	}
	c := makeUserController(repo)
	handler, token := withClaims(t, http.HandlerFunc(c.Update), "user", "user-123")

	req := httptest.NewRequest(http.MethodPut, "/users/user-123", jsonBody(t, map[string]string{
		"first_name": "",
		"last_name":  "",
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "user-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- Delete ---

func TestDelete_OwnAccount(t *testing.T) {
	repo := &mockUserRepo{
		DeleteFn: func(ctx context.Context, id string) error { return nil },
	}
	c := makeUserController(repo)
	handler, token := withClaims(t, http.HandlerFunc(c.Delete), "user", "user-123")

	req := httptest.NewRequest(http.MethodDelete, "/users/user-123", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "user-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestDelete_NoAuth(t *testing.T) {
	c := makeUserController(&mockUserRepo{})

	req := httptest.NewRequest(http.MethodDelete, "/users/user-123", nil)
	req.SetPathValue("id", "user-123")
	rr := httptest.NewRecorder()

	c.Delete(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestDelete_WrongAccount(t *testing.T) {
	c := makeUserController(&mockUserRepo{})
	handler, token := withClaims(t, http.HandlerFunc(c.Delete), "user", "other-user-456")

	req := httptest.NewRequest(http.MethodDelete, "/users/user-123", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "user-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestDelete_NotFound(t *testing.T) {
	repo := &mockUserRepo{
		DeleteFn: func(ctx context.Context, id string) error { return userDomain.ErrUserNotFound },
	}
	c := makeUserController(repo)
	handler, token := withClaims(t, http.HandlerFunc(c.Delete), "user", "user-123")

	req := httptest.NewRequest(http.MethodDelete, "/users/user-123", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", "user-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

