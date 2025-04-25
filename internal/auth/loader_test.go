package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jeanmolossi/verbose-adventure/internal/config"
)

func encryptAESGCM(t *testing.T, key []byte, plaintext string) []byte {
	t.Helper()

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("failed to create cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("failed to create GCM: %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		t.Fatalf("failed to read nonce: %v", err)
	}

	return append(nonce, gcm.Seal(nil, nonce, []byte(plaintext), nil)...)
}

func TestDecryptSecret_Success(t *testing.T) {
	rawKey := make([]byte, 32)
	if _, err := rand.Read(rawKey); err != nil {
		t.Fatal(err)
	}

	keyB64 := base64.StdEncoding.EncodeToString(rawKey)

	ciphertext := encryptAESGCM(t, rawKey, "super-secret")

	got, err := decryptSecret(ciphertext, keyB64)
	if err != nil {
		t.Fatal(err)
	}

	if got != "super-secret" {
		t.Fatalf("want %q, got %q", "super-secret", got)
	}
}

func TestDecryptSecret_InvalidBase64(t *testing.T) {
	ciphertext := []byte("data")

	if _, err := decryptSecret(ciphertext, "bad-base64"); err == nil {
		t.Fatal("expected error on base64")
	}
}

func TestDecryptSecret_InvalidKeyLength(t *testing.T) {
	// valid key length to encrypt
	rawKey := make([]byte, 32)
	if _, err := rand.Read(rawKey); err != nil {
		t.Fatalf("failed to randomize key: %v", err)
	}
	ciphertext := encryptAESGCM(t, rawKey, "x")

	// invalid key length
	shortKey := make([]byte, 10)
	if _, err := rand.Read(shortKey); err != nil {
		t.Fatalf("failed to randomize short key: %v", err)
	}

	keyB64 := base64.StdEncoding.EncodeToString(shortKey)
	_, err := decryptSecret(ciphertext, keyB64)
	if err == nil {
		t.Error("expected error for invalid key length, got nil")
	}
}

func TestDecryptSecret_CipherTooShort(t *testing.T) {
	rawKey := make([]byte, 32)
	if _, err := rand.Read(rawKey); err != nil {
		t.Fatalf("failed to randomize key: %v", err)
	}

	keyB64 := base64.StdEncoding.EncodeToString(rawKey)

	short := make([]byte, 0)
	_, err := decryptSecret(short, keyB64)
	if err == nil {
		t.Error("expected error for invalid key length, got nil")
	}
}

func TestLoadIdentityProviders_Success(t *testing.T) {
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"issuer":"` + ts.URL + `","jwks_uri":"` + ts.URL + `/jwks"}`))
		case "/jwks":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"keys":[]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}

	defer db.Close()

	rawKey := make([]byte, 32)
	if _, err := rand.Read(rawKey); err != nil {
		t.Fatalf("failed to randomize key: %v", err)
	}
	keyB64 := base64.StdEncoding.EncodeToString(rawKey)
	cfg := &config.Config{EncryptionKey: keyB64, BaseURL: ts.URL}

	secretPlain := "secret"
	ciphertext := encryptAESGCM(t, rawKey, secretPlain)

	rows := sqlmock.NewRows([]string{"id", "tenant_id", "type", "metadata_url", "client_id", "client_secret_enc", "enabled"}).
		AddRow(1, 42, "oidc", ts.URL, "client-id", ciphertext, true)

	mock.ExpectQuery("SELECT id, tenant_id, type, metadata_url, client_id, client_secret_enc, enabled FROM identity_providers WHERE enabled = ?").
		WillReturnRows(rows)

	providers, err := LoadIdentityProviders(cfg, db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(providers) != 1 {
		t.Errorf("expected 1 providers, got %d", len(providers))
	}

	p := providers[0]
	if p.TenantID() != 42 {
		t.Errorf("expected tenantID 42, got %d", p.TenantID())
	}

	if p.Type() != "oidc" {
		t.Errorf("expected type oidc, got %s", p.Type())
	}
}

func TestLoadIdentityProviders_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}

	defer db.Close()

	cfg := &config.Config{EncryptionKey: "", BaseURL: ""}

	mock.ExpectQuery("SELECT id, tenant_id, type, metadata_url, client_id, client_secret_enc, enabled FROM identity_providers WHERE enabled = ?").
		WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id", "type", "metadata_url", "client_id", "client_secret_enc", "enabled"}))

	providers, err := LoadIdentityProviders(cfg, db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(providers) != 0 {
		t.Errorf("expected 0 providers, got %d", len(providers))
	}
}

func TestOIDCProviderCallback_NoCode(t *testing.T) {
	// Prepare provider
	rawKey := make([]byte, 32)
	if _, err := rand.Read(rawKey); err != nil {
		t.Fatalf("failed to randomize key: %v", err)
	}

	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"issuer":"` + ts.URL + `","jwks_uri":"` + ts.URL + `/jwks"}`))
		case "/jwks":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"keys":[]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	rec := idpRecord{
		ID:              1,
		TenantID:        7,
		ProviderType:    "oidc",
		MetadataURL:     ts.URL,
		ClientID:        "id",
		ClientSecretEnc: encryptAESGCM(t, rawKey, "secret"),
		Enabled:         true,
	}

	cfg := &config.Config{EncryptionKey: base64.StdEncoding.EncodeToString(rawKey), BaseURL: ts.URL}
	pIface, err := newOIDCProvider(context.Background(), rec, "secret", cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}
	op := pIface.(*oidcProvider)

	req := httptest.NewRequest(http.MethodGet, "/callback", nil)
	if _, err := op.Callback(context.Background(), req); err == nil {
		t.Error("expected error for missing code, got nil")
	}
}

func TestOIDCProviderCallback_InvalidCode(t *testing.T) {
	// Similar setup
	rawKey := make([]byte, 32)
	if _, err := rand.Read(rawKey); err != nil {
		t.Fatalf("failed to randomize key: %v", err)
	}
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"issuer":"` + ts.URL + `","jwks_uri":"` + ts.URL + `/jwks","token_endpoint":"` + ts.URL + `/token"}`))
		case "/jwks":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"keys":[]}`))
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"access_token":"x"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	rec := idpRecord{
		ID:              1,
		TenantID:        7,
		ProviderType:    "oidc",
		MetadataURL:     ts.URL,
		ClientID:        "id",
		ClientSecretEnc: encryptAESGCM(t, rawKey, "secret"),
		Enabled:         true,
	}

	cfg := &config.Config{EncryptionKey: base64.StdEncoding.EncodeToString(rawKey), BaseURL: ts.URL}
	pIface, err := newOIDCProvider(context.Background(), rec, "secret", cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}
	op := pIface.(*oidcProvider)

	req := httptest.NewRequest(http.MethodGet, "/callback?code=foo", nil)
	if _, err := op.Callback(context.Background(), req); err == nil {
		t.Error("expected error for invalid code exchange, got nil")
	}
}
