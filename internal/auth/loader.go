package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"

	"github.com/coreos/go-oidc"
	"github.com/jeanmolossi/verbose-adventure/internal/config"
	"golang.org/x/oauth2"
)

type IdentityProvider interface {
	TenantID() int64
	Type() string
	AuthURL(state string) string
	Callback(ctx context.Context, req *http.Request) (*AuthResult, error)
}

type AuthResult struct {
	TenantID   int64
	UserID     string
	Email      string
	RawIDToken *oidc.IDToken
}

type idpRecord struct {
	ID              int64
	TenantID        int64
	ProviderType    string
	MetadataURL     string
	ClientID        string
	ClientSecretEnc []byte
	Enabled         bool
}

func LoadIdentityProviders(cfg *config.Config, db *sql.DB) ([]IdentityProvider, error) {
	ctx := context.Background()

	query := `
    SELECT id, tenant_id, type, metadata_url, client_id, client_secret_enc, enabled
    FROM identity_providers
    WHERE enabled = ?
    `

	rows, err := db.QueryContext(ctx, query, 1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	providers := make([]IdentityProvider, 0, 2) // pre-alloc 2 providers space
	for rows.Next() {
		var rec idpRecord
		if err := rows.Scan(
			&rec.ID,
			&rec.TenantID,
			&rec.ProviderType,
			&rec.MetadataURL,
			&rec.ClientID,
			&rec.ClientSecretEnc,
			&rec.Enabled,
		); err != nil {
			return nil, err
		}

		secret, err := decryptSecret(rec.ClientSecretEnc, cfg.EncryptionKey)
		if err != nil {
			return nil, err
		}

		switch rec.ProviderType {
		case "oidc":
			p, err := newOIDCProvider(ctx, rec, secret, cfg)
			if err != nil {
				return nil, err
			}

			providers = append(providers, p)
		case "saml":
			// TODO: implement newSAMLProvider
		default:
			// ignore
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return providers, nil
}

// decryptSecret decipher the client_secret using AES-GCM
func decryptSecret(ciphertext []byte, keyBase64 string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return "", err
	}

	if len(key) != 32 {
		return "", errors.New("invalid encryption key length")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", nil
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func newOIDCProvider(ctx context.Context, rec idpRecord, secret string, cfg *config.Config) (IdentityProvider, error) {
	provider, err := oidc.NewProvider(ctx, rec.MetadataURL)
	if err != nil {
		return nil, err
	}

	oauth2Cfg := &oauth2.Config{
		ClientID:     rec.ClientID,
		ClientSecret: secret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  cfg.BaseURL + "/" + strconv.FormatInt(rec.TenantID, 10) + "/oidc/callback",
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: rec.ClientID})
	return &oidcProvider{
		id:       rec.ID,
		tenantID: rec.TenantID,
		oauth2:   oauth2Cfg,
		verifier: verifier,
	}, nil
}

type oidcProvider struct {
	id       int64
	tenantID int64
	oauth2   *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func (o *oidcProvider) TenantID() int64 { return o.tenantID }
func (o *oidcProvider) Type() string    { return "oidc" }
func (o *oidcProvider) AuthURL(state string) string {
	return o.oauth2.AuthCodeURL(state)
}

func (o *oidcProvider) Callback(ctx context.Context, req *http.Request) (*AuthResult, error) {
	q := req.URL.Query()
	code := q.Get("code")
	if code == "" {
		return nil, errors.New("authorization code not found")
	}

	tok, err := o.oauth2.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	rawID, ok := tok.Extra("id_token").(string)
	if !ok || rawID == "" {
		return nil, errors.New("id_token not found in token response")
	}

	idTok, err := o.verifier.Verify(ctx, rawID)
	if err != nil {
		return nil, err
	}

	var claims struct {
		Email   string `json:"email"`
		Subject string `json:"sub"`
	}

	if err := idTok.Claims(&claims); err != nil {
		return nil, err
	}

	return &AuthResult{
		TenantID:   o.tenantID,
		UserID:     claims.Subject,
		Email:      claims.Email,
		RawIDToken: idTok,
	}, nil
}
