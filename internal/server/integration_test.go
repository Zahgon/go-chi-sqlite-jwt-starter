package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-chi-sqlite-jwt-starter/config"
	"go-chi-sqlite-jwt-starter/internal/database"
	"go-chi-sqlite-jwt-starter/internal/provider"
	"go-chi-sqlite-jwt-starter/internal/server"
)

// testEnv wires up config, database and provider against a temp DB folder so the
// full HTTP stack can be exercised without a .env file. It is intentionally
// framework-agnostic: it drives the router returned by server.Initialize()
// purely over HTTP, so it passes unchanged on both the original and migrated
// implementations.
func setup(t *testing.T) http.Handler {
	t.Helper()

	config.Variables.Port = "0"
	config.Variables.DB_FOLDER = t.TempDir()
	config.Variables.AUTH_PRIVATE_KEY = "test-secret-key"

	database.Initialize()
	provider.Initialize()
	return server.Initialize()
}

func doJSON(t *testing.T, h http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHealthEndpoint(t *testing.T) {
	h := setup(t)
	rec := doJSON(t, h, http.MethodGet, "/health", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("health status = %d, want 200", rec.Code)
	}
}

func TestLoginUnknownUserUnauthorized(t *testing.T) {
	h := setup(t)
	rec := doJSON(t, h, http.MethodPost, "/auth/login", "", map[string]string{
		"username": "nobody",
		"password": "whatever",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("login unknown user status = %d, want 401", rec.Code)
	}
}

func TestRegisterInvalidPayload(t *testing.T) {
	h := setup(t)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("{not json"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("register invalid payload status = %d, want 400", rec.Code)
	}
}

func TestProtectedRouteWithoutTokenIsRejected(t *testing.T) {
	h := setup(t)
	rec := doJSON(t, h, http.MethodGet, "/category/list", "", nil)
	if rec.Code == http.StatusOK {
		t.Fatalf("protected route without token returned 200, want non-200")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("protected route without token status = %d, want 401", rec.Code)
	}
}

func TestAdminRouteWithoutTokenIsRejected(t *testing.T) {
	h := setup(t)
	rec := doJSON(t, h, http.MethodGet, "/admin/test-token", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("admin route without token status = %d, want 401", rec.Code)
	}
}
