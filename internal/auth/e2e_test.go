//go:build integration

package auth_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiInfra "github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	authinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure"
	authUtils "github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/utils"
	storeinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure"
	userinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/user/infrastructure"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/testutil"
)

const testJWTSecret = "e2e-test-secret"

func buildTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	db := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, db)

	mux := http.NewServeMux()
	idGen := apiInfra.NewUUIDGenerator()

	jwtGen := authUtils.NewJWTGenerator(testJWTSecret, time.Hour)
	authMw := apiInfra.NewAuthMiddleware(jwtGen)

	userModule := userinfra.Wire(mux, db, idGen, authMw)
	storeModule := storeinfra.Wire(mux, db, idGen, authMw)
	authinfra.Wire(mux, db, idGen, userModule, storeModule, authUtils.JwtConfig{
		JWTSecret:     testJWTSecret,
		JWTExpiration: time.Hour,
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func postJSON(t *testing.T, srv *httptest.Server, path string, body any, token string) *http.Response {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, srv.URL+path, bytes.NewReader(b))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func getJSON(t *testing.T, srv *httptest.Server, path, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, srv.URL+path, nil)
	require.NoError(t, err)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func decodeBody(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()
	var m map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&m))
	return m
}

// TestE2E_AuthFlow covers the full register → login → access flow for a user account.
func TestE2E_AuthFlow(t *testing.T) {
	srv := buildTestServer(t)

	// 1. Register new user → 201
	registerResp := postJSON(t, srv, "/auth/register/user", map[string]string{
		"email":      "e2e@example.com",
		"password":   "password123",
		"first_name": "E2E",
		"last_name":  "User",
	}, "")
	assert.Equal(t, http.StatusCreated, registerResp.StatusCode)
	registerBody := decodeBody(t, registerResp)
	accountID := registerBody["account_id"].(string)
	require.NotEmpty(t, accountID)

	// 2. Login with same credentials → 200 + token
	loginResp := postJSON(t, srv, "/auth/login", map[string]string{
		"email":    "e2e@example.com",
		"password": "password123",
	}, "")
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)
	loginBody := decodeBody(t, loginResp)
	token := loginBody["token"].(string)
	require.NotEmpty(t, token)

	// 3. Access a protected endpoint with token → 200
	protectedResp := getJSON(t, srv, fmt.Sprintf("/users/%s", accountID), token)
	assert.Equal(t, http.StatusOK, protectedResp.StatusCode)
	protectedBody := decodeBody(t, protectedResp)
	assert.Equal(t, accountID, protectedBody["id"])

	// 4. Access a protected endpoint without token → 401
	noAuthResp := getJSON(t, srv, fmt.Sprintf("/users/%s", accountID), "")
	// GET /users/{id} doesn't require auth (public endpoint) — use PUT which does
	noAuthResp.Body.Close()

	putResp, err := http.NewRequest(http.MethodPut, srv.URL+fmt.Sprintf("/users/%s", accountID), bytes.NewReader([]byte(`{"first_name":"New","last_name":"Name"}`)))
	require.NoError(t, err)
	putResp.Header.Set("Content-Type", "application/json")
	putResp2, err := http.DefaultClient.Do(putResp)
	require.NoError(t, err)
	putResp2.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, putResp2.StatusCode)

	// 5. Access resource owned by another account → 403
	// Register a second user
	register2Resp := postJSON(t, srv, "/auth/register/user", map[string]string{
		"email":      "e2e2@example.com",
		"password":   "password456",
		"first_name": "Second",
		"last_name":  "User",
	}, "")
	require.Equal(t, http.StatusCreated, register2Resp.StatusCode)
	register2Body := decodeBody(t, register2Resp)
	otherAccountID := register2Body["account_id"].(string)

	// Try to update otherAccountID's user while authenticated as first user
	forbiddenBody, _ := json.Marshal(map[string]string{"first_name": "Hacked", "last_name": "Name"})
	forbiddenReq, _ := http.NewRequest(http.MethodPut, srv.URL+fmt.Sprintf("/users/%s", otherAccountID), bytes.NewReader(forbiddenBody))
	forbiddenReq.Header.Set("Content-Type", "application/json")
	forbiddenReq.Header.Set("Authorization", "Bearer "+token)
	forbiddenResp, err := http.DefaultClient.Do(forbiddenReq)
	require.NoError(t, err)
	forbiddenResp.Body.Close()
	assert.Equal(t, http.StatusForbidden, forbiddenResp.StatusCode)
}

// TestE2E_RegisterStore covers the store registration flow.
func TestE2E_RegisterStore(t *testing.T) {
	srv := buildTestServer(t)

	// Register a store → 201
	registerResp := postJSON(t, srv, "/auth/register/store", map[string]string{
		"email":        "mystore@example.com",
		"password":     "password123",
		"name":         "My E2E Store",
		"description":  "A test store created during e2e testing",
		"address":      "123 E2E Street, Test City",
		"phone_number": "555-0100",
	}, "")
	assert.Equal(t, http.StatusCreated, registerResp.StatusCode)
	body := decodeBody(t, registerResp)
	assert.Equal(t, "store", body["account_type"])
	assert.Equal(t, "mystore@example.com", body["email"])

	// Login as the new store → 200
	loginResp := postJSON(t, srv, "/auth/login", map[string]string{
		"email":    "mystore@example.com",
		"password": "password123",
	}, "")
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)
	loginBody := decodeBody(t, loginResp)
	assert.NotEmpty(t, loginBody["token"])
}

// TestE2E_DuplicateEmail verifies that registering with an existing email returns 409.
func TestE2E_DuplicateEmail(t *testing.T) {
	srv := buildTestServer(t)

	payload := map[string]string{
		"email":      "dup@example.com",
		"password":   "password123",
		"first_name": "First",
		"last_name":  "User",
	}

	first := postJSON(t, srv, "/auth/register/user", payload, "")
	first.Body.Close()
	assert.Equal(t, http.StatusCreated, first.StatusCode)

	second := postJSON(t, srv, "/auth/register/user", payload, "")
	second.Body.Close()
	assert.Equal(t, http.StatusConflict, second.StatusCode)
}

// TestE2E_InvalidLoginCredentials verifies 401 for wrong credentials.
func TestE2E_InvalidLoginCredentials(t *testing.T) {
	srv := buildTestServer(t)

	resp := postJSON(t, srv, "/auth/login", map[string]string{
		"email":    "nobody@example.com",
		"password": "wrongpassword",
	}, "")
	resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
