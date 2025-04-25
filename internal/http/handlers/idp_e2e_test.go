package handlers_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jeanmolossi/verbose-adventure/internal/config"
	h "github.com/jeanmolossi/verbose-adventure/internal/http/handlers"
	"github.com/jeanmolossi/verbose-adventure/internal/repo"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

// fakeRepo implements IdentityProviderRepository for testing
type fakeRepo struct {
	list     []*repo.IdentityProviderRecord
	create   *repo.IdentityProviderRecord
	update   *repo.IdentityProviderRecord
	deleteID int64

	// errors to return
	listErr   error
	getErr    error
	createErr error
	updateErr error
	deleteErr error
	getRec    *repo.IdentityProviderRecord
}

var _ repo.IdentityProviderRepository = (*fakeRepo)(nil)

func (f *fakeRepo) ListByTenant(ctx context.Context, tenantID int64) ([]*repo.IdentityProviderRecord, error) {
	return f.list, f.listErr
}

func (f *fakeRepo) GetByID(ctx context.Context, id int64) (*repo.IdentityProviderRecord, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.getRec, nil
}

func (f *fakeRepo) Create(ctx context.Context, rec *repo.IdentityProviderRecord) (int64, error) {
	f.create = rec
	return 123, f.createErr
}

func (f *fakeRepo) Update(ctx context.Context, rec *repo.IdentityProviderRecord) error {
	f.update = rec
	return f.updateErr
}

func (f *fakeRepo) Delete(ctx context.Context, id int64) error {
	f.deleteID = id
	return f.deleteErr
}

func setup() (*echo.Echo, *fakeRepo) {
	e := echo.New()

	repo := &fakeRepo{}

	cfg := &config.Config{
		EncryptionKey: base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901")),
	}

	handler := h.NewIDPHandler(h.IDPHandlerParams{Repo: repo, Cfg: cfg})
	handler.Register(e)

	return e, repo
}

func TestList_Success(t *testing.T) {
	e, repoMock := setup()
	// prepare fake record
	secret := []byte("secret")
	repoMock.list = []*repo.IdentityProviderRecord{{
		ID: 1, TenantID: 99, ProviderType: "oidc", MetadataURL: "url",
		ClientID: "cid", ClientSecretEnc: secret, Enabled: true,
		CreatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC),
	}}

	req := httptest.NewRequest(http.MethodGet, "/admin/tenants/99/idps", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	reqRes := rec.Result()
	defer reqRes.Body.Close()
	require.Equal(t, http.StatusOK, reqRes.StatusCode)

	var out []h.IdpResponse
	err := json.NewDecoder(reqRes.Body).Decode(&out)
	require.NoError(t, err)
	require.Len(t, out, 1)

	reqResp := out[0]
	require.Equal(t, int64(1), reqResp.ID)
	require.Equal(t, int64(99), reqResp.TenantID)
	require.Equal(t, "oidc", reqResp.ProviderType)
}

func TestGet_NotFound(t *testing.T) {
	e, repo := setup()
	repo.getRec = nil

	req := httptest.NewRequest(http.MethodGet, "/admin/tenants/1/idps/42", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	reqRes := rec.Result()
	defer reqRes.Body.Close()
	require.Equal(t, http.StatusNotFound, reqRes.StatusCode)
}

func TestCreate_Success(t *testing.T) {
	e, repo := setup()
	// build request payload
	payload := map[string]interface{}{
		"type": "oidc", "metadata_url": "https://example.com",
		"client_id": "cid", "client_secret": "secret", "enabled": true,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/admin/tenants/5/idps", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	reqRes := rec.Result()
	defer reqRes.Body.Close()
	require.Equal(t, http.StatusCreated, reqRes.StatusCode)

	// Check repo.create was called
	require.NotNil(t, repo.create)
	require.Equal(t, int64(5), repo.create.TenantID)
	require.Equal(t, "oidc", repo.create.ProviderType)
}

func TestUpdate_Success(t *testing.T) {
	e, repo := setup()
	// build request payload
	payload := map[string]interface{}{
		"type": "saml", "metadata_url": "https://saml",
		"client_id": "cid2", "client_secret": "secret2", "enabled": false,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/admin/tenants/7/idps/77", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	reqRes := rec.Result()
	defer reqRes.Body.Close()
	require.Equal(t, http.StatusNoContent, reqRes.StatusCode)

	// Check repo.update was called
	require.NotNil(t, repo.update)
	require.Equal(t, int64(77), repo.update.ID)
	require.Equal(t, "saml", repo.update.ProviderType)
}

func TestDelete_Success(t *testing.T) {
	e, repo := setup()

	req := httptest.NewRequest(http.MethodDelete, "/admin/tenants/9/idps/99", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	reqRes := rec.Result()
	defer reqRes.Body.Close()
	require.Equal(t, http.StatusNoContent, reqRes.StatusCode)
	require.Equal(t, int64(99), repo.deleteID)
}
