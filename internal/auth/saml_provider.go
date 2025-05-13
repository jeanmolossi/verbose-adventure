package auth

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"

	"github.com/crewjam/saml/samlsp"
	"github.com/jeanmolossi/verbose-adventure/internal/config"
)

type samlProvider struct {
	m         *samlsp.Middleware
	tentantID int64
}

func (s *samlProvider) TenantID() int64 { return s.tentantID }
func (s *samlProvider) Type() string    { return "saml" }

// AuthURL inicia o fluxo SAML (AuthnRequest → RedirectBinding)
func (s *samlProvider) AuthURL(state string) string {
	// fake HTTPS request to capture redirect location
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// injects relay state
	q := req.URL.Query()
	q.Set("RelayState", state)
	req.URL.RawQuery = q.Encode()

	rec := httptest.NewRecorder()
	s.m.HandleStartAuthFlow(rec, req)
	return rec.Header().Get("Location")
}

func (s *samlProvider) Callback(ctx context.Context, r *http.Request) (*AuthResult, error) {
	// simulate a response writer
	rec := httptest.NewRecorder()
	s.m.ServeACS(rec, r)
	if rec.Code != http.StatusOK {
		return nil, errors.New("SAML ACS error: " + rec.Body.String())
	}

	// extract session and attrs
	session := samlsp.SessionFromContext(r.Context())
	sessWithAttrs, ok := session.(samlsp.SessionWithAttributes)
	if !ok {
		return nil, errors.New("session does not have attributes")
	}

	attrs := sessWithAttrs.GetAttributes()

	email := attrs.Get("email")
	userID := attrs.Get("uid")

	return &AuthResult{
		TenantID: s.tentantID,
		UserID:   userID,
		Email:    email,
	}, nil
}

func newSAMLProvider(ctx context.Context, rec idpRecord, cfg *config.Config) (IdentityProvider, error) {
	keyPair, err := tls.LoadX509KeyPair(cfg.SAMLCertPath, cfg.SAMLKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load key pairs: %w", err)
	}

	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	idpMetadataURL, err := url.Parse(rec.MetadataURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metadataURL: %w", err)
	}

	fmt.Println("REC METAURL", rec.MetadataURL)

	idpMetadata, err := samlsp.FetchMetadata(ctx, http.DefaultClient, *idpMetadataURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metadata: %w", err)
	}

	parsedUrl, err := url.Parse(cfg.BaseURL + "/" + strconv.FormatInt(rec.TenantID, 10))
	if err != nil {
		return nil, err
	}

	spOptions := samlsp.Options{
		EntityID:    rec.ClientID,
		URL:         *parsedUrl,
		Key:         keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate: keyPair.Leaf,
		IDPMetadata: idpMetadata,
		SignRequest: true,
		RelayStateFunc: func(w http.ResponseWriter, r *http.Request) string {
			return r.URL.Query().Get("RelayState")
		},
	}

	log.Printf("SAML ACS URL = %q", spOptions.URL.String()+"/saml/callback")

	m, err := samlsp.New(spOptions)
	if err != nil {
		return nil, err
	}

	return &samlProvider{
		m:         m,
		tentantID: rec.TenantID,
	}, nil
}
