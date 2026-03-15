package infrastructure_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/utils"
)

func newTestJWTGen() *utils.JWTGenerator {
	return utils.NewJWTGenerator("test-secret", time.Hour)
}

func TestMiddleware_ValidToken(t *testing.T) {
	jwtGen := newTestJWTGen()
	token, err := jwtGen.Generate("auth-1", domain.AccountTypeUser, "acc-1")
	require.NoError(t, err)

	mw := infrastructure.NewAuthMiddleware(jwtGen)
	called := false
	handler := mw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		claims, ok := infrastructure.GetClaims(r.Context())
		assert.True(t, ok)
		assert.Equal(t, "auth-1", claims.AuthID)
		assert.Equal(t, "acc-1", claims.AccountID)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestMiddleware_MissingAuthHeader(t *testing.T) {
	mw := infrastructure.NewAuthMiddleware(newTestJWTGen())
	handler := mw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestMiddleware_InvalidToken(t *testing.T) {
	mw := infrastructure.NewAuthMiddleware(newTestJWTGen())
	handler := mw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestMiddleware_MalformedBearerFormat(t *testing.T) {
	mw := infrastructure.NewAuthMiddleware(newTestJWTGen())
	handler := mw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Token sometoken")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestGetClaims_Present(t *testing.T) {
	jwtGen := newTestJWTGen()
	token, err := jwtGen.Generate("auth-1", domain.AccountTypeStore, "acc-42")
	require.NoError(t, err)

	mw := infrastructure.NewAuthMiddleware(jwtGen)
	var gotClaims *utils.Claims
	var gotOk bool
	handler := mw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotClaims, gotOk = infrastructure.GetClaims(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	assert.True(t, gotOk)
	require.NotNil(t, gotClaims)
	assert.Equal(t, "acc-42", gotClaims.AccountID)
}

func TestGetClaims_Absent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	claims, ok := infrastructure.GetClaims(req.Context())
	assert.False(t, ok)
	assert.Nil(t, claims)
}
