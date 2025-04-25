package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/jeanmolossi/verbose-adventure/internal/auth"
	"github.com/jeanmolossi/verbose-adventure/internal/config"
	"github.com/jeanmolossi/verbose-adventure/internal/http/handlers"
)

type fakeOIDC struct {
	tenant int64
}

func (f *fakeOIDC) TenantID() int64             { return f.tenant }
func (f *fakeOIDC) Type() string                { return "oidc" }
func (f *fakeOIDC) AuthURL(state string) string { return "https://fakeidp.com/auth?state=" + state }
func (f *fakeOIDC) Callback(ctx context.Context, req *http.Request) (*auth.AuthResult, error) {
	// Ignore code, return fixed AuthResult
	return &auth.AuthResult{
		TenantID: f.tenant,
		UserID:   "user123",
		Email:    "user@example.com",
	}, nil
}

func setupServer(t *testing.T, providers []auth.IdentityProvider, cfg *config.Config) *echo.Echo {
	e := echo.New()
	h := handlers.NewAuthHandler(providers, cfg)
	h.Register(e)
	return e
}

func TestAuthEndToEnd_LoginAndCallback(t *testing.T) {
	// Configuration
	cfg := &config.Config{JWTSecret: "test-secret"}
	provider := &fakeOIDC{tenant: 99}
	e := setupServer(t, []auth.IdentityProvider{provider}, cfg)

	// Test Login
	req := httptest.NewRequest(http.MethodGet, "/99/oidc/login", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	reqed := rec.Result()
	reqed.Body.Close()

	require.Equal(t, http.StatusFound, reqed.StatusCode)
	loc, err := reqed.Location()
	require.NoError(t, err)
	require.Contains(t, loc.String(), "https://fakeidp.com/auth?state=")

	// Simulate callback
	callbackReq := httptest.NewRequest(http.MethodGet, "/99/oidc/callback?code=irrelevant", nil)
	callbackRec := httptest.NewRecorder()
	e.ServeHTTP(callbackRec, callbackReq)
	res := callbackRec.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)

	var body map[string]string
	err = json.NewDecoder(res.Body).Decode(&body)
	require.NoError(t, err)
	tokenString, ok := body["token"]
	require.True(t, ok, "response should contain token field")

	// Validate JWT
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			t.Fatalf("unexpected signing method: %v", token.Method)
		}
		return []byte(cfg.JWTSecret), nil
	})
	require.NoError(t, err)
	require.True(t, token.Valid)

	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok)
	require.Equal(t, float64(99), claims["tenant_id"])
	require.Equal(t, "user123", claims["user_id"])
	require.Equal(t, "user@example.com", claims["email"])
}
