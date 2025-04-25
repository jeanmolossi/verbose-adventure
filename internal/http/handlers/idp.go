package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/jeanmolossi/verbose-adventure/internal/config"
	"github.com/jeanmolossi/verbose-adventure/internal/repo"
	"github.com/jeanmolossi/verbose-adventure/pkg/stripslash"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

// IDPHandler gerencia CRUD de Identity Providers
type IDPHandler struct {
	repo repo.IdentityProviderRepository
	cfg  *config.Config
}

type IDPHandlerParams struct {
	fx.In
	Repo repo.IdentityProviderRepository
	Cfg  *config.Config
}

// NewIDPHandler cria um novo handler, injetando o repo
func NewIDPHandler(p IDPHandlerParams) *IDPHandler {
	return &IDPHandler{repo: p.Repo, cfg: p.Cfg}
}

// idpRequest representa o payload de criação/atualização
type idpRequest struct {
	ProviderType string `json:"type" validate:"required,oneof=oidc saml"`
	MetadataURL  string `json:"metadata_url" validate:"required,url"`
	ClientID     string `json:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" validate:"required"`
	Enabled      bool   `json:"enabled"`
}

// IdpResponse representa a resposta ao cliente
type IdpResponse struct {
	ID              int64  `json:"id"`
	TenantID        int64  `json:"tenant_id"`
	ProviderType    string `json:"type"`
	MetadataURL     string `json:"metadata_url"`
	ClientSecretEnc string `json:"client_secret_enc"`
	Enabled         bool   `json:"enabled"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// Register associa as rotas de administração de IdP
func (h *IDPHandler) Register(e *echo.Echo) {
	g := e.Group("/admin/tenants/:tenantID/idps")
	g.GET(":id", h.Get)
	g.GET("", h.List)
	g.POST("", h.Create)
	g.PUT(":id", h.Update)
	g.DELETE(":id", h.Delete)
}

// List retorna todos os providers de um tenant
func (h *IDPHandler) List(c echo.Context) error {
	tid, err := parseTenant(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "tenant is not valid")
	}

	recs, err := h.repo.ListByTenant(c.Request().Context(), tid)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var out []IdpResponse
	for _, r := range recs {
		out = append(out, IdpResponse{
			ID:              r.ID,
			TenantID:        r.TenantID,
			ProviderType:    r.ProviderType,
			MetadataURL:     r.MetadataURL,
			ClientSecretEnc: base64.StdEncoding.EncodeToString(r.ClientSecretEnc),
			Enabled:         r.Enabled,
			CreatedAt:       r.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       r.UpdatedAt.Format(time.RFC3339),
		})
	}

	return c.JSON(http.StatusOK, out)
}

// Get retorna um provider específico
func (h *IDPHandler) Get(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, c.Param("id"))
	}

	rec, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if rec == nil {
		return c.NoContent(http.StatusNotFound)
	}

	resp := IdpResponse{
		ID:              rec.ID,
		TenantID:        rec.TenantID,
		ProviderType:    rec.ProviderType,
		MetadataURL:     rec.MetadataURL,
		ClientSecretEnc: base64.StdEncoding.EncodeToString(rec.ClientSecretEnc),
		Enabled:         rec.Enabled,
		CreatedAt:       rec.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       rec.UpdatedAt.Format(time.RFC3339),
	}

	return c.JSON(http.StatusOK, resp)
}

// Create adiciona um novo Identity Provider
func (h *IDPHandler) Create(c echo.Context) error {
	var req idpRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// decodifica secret
	rawSecret := []byte(req.ClientSecret)
	secret, err := encryptSecret(rawSecret, h.cfg.EncryptionKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid client_secret_enc, must be base64")
	}

	tenant, err := parseTenant(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tenant id")
	}

	rec := &repo.IdentityProviderRecord{
		TenantID:        tenant,
		ProviderType:    req.ProviderType,
		MetadataURL:     req.MetadataURL,
		ClientID:        req.ClientID,
		ClientSecretEnc: secret,
		Enabled:         req.Enabled,
	}

	id, err := h.repo.Create(c.Request().Context(), rec)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	rec.ID = id

	return c.JSON(http.StatusCreated, map[string]int64{"id": id})
}

// Update modifica um Identity Provider existente
func (h *IDPHandler) Update(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	var req idpRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	rawSecret := []byte(req.ClientSecret)
	secret, err := encryptSecret(rawSecret, h.cfg.EncryptionKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to encrypt secret")
	}

	tenant, err := parseTenant(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tenant id")
	}

	rec := &repo.IdentityProviderRecord{
		ID:              id,
		TenantID:        tenant,
		ProviderType:    req.ProviderType,
		MetadataURL:     req.MetadataURL,
		ClientID:        req.ClientID,
		ClientSecretEnc: secret,
		Enabled:         req.Enabled,
	}
	if err := h.repo.Update(c.Request().Context(), rec); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// Delete remove um Identity Provider
func (h *IDPHandler) Delete(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	if err := h.repo.Delete(c.Request().Context(), id); err != nil {
		if err == sql.ErrNoRows {
			return c.NoContent(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// parseTenant obtém tenantID do context
func parseTenant(c echo.Context) (int64, error) {
	tenant := stripslash.ParamWithoutAnySlashes(c, "tenantID")
	return strconv.ParseInt(tenant, 10, 64)
}

func parseID(c echo.Context) (int64, error) {
	pid := stripslash.ParamWithoutBackslash(c, "id")
	return strconv.ParseInt(pid, 10, 64)
}

// encryptSecret aplica AES-GCM com chave base64 (32 bytes) e retorna nonce|ciphertext
func encryptSecret(plaintext []byte, keyB64 string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return append(nonce, gcm.Seal(nil, nonce, plaintext, nil)...), nil
}
