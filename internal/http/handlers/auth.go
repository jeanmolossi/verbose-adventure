package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jeanmolossi/verbose-adventure/internal/auth"
	"github.com/jeanmolossi/verbose-adventure/internal/config"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	Providers []auth.IdentityProvider
	Config    *config.Config
}

func NewAuthHandler(providers []auth.IdentityProvider, cfg *config.Config) *AuthHandler {
	return &AuthHandler{Providers: providers, Config: cfg}
}

func (h *AuthHandler) getIdentityProvider(c echo.Context) (auth.IdentityProvider, error) {
	tenantID, err := strconv.ParseInt(c.Param("tenantID"), 10, 64)
	if err != nil {
		return nil, c.JSON(http.StatusBadRequest, echo.Map{"message": "invalid tenantID"})
	}

	typeID := c.Param("idpType")

	var p auth.IdentityProvider
	for _, pr := range h.Providers {
		if pr.TenantID() == tenantID && pr.Type() == typeID {
			p = pr
			break
		}
	}

	if p == nil {
		return nil, c.JSON(http.StatusNotFound, echo.Map{"message": "provider not found"})
	}

	return p, nil
}

func (h *AuthHandler) Login(c echo.Context) error {
	p, err := h.getIdentityProvider(c)
	if err != nil {
		return err
	}

	state := strconv.FormatInt(time.Now().UnixNano(), 10)
	url := p.AuthURL(state)
	return c.Redirect(http.StatusFound, url)
}

func (h *AuthHandler) Callback(c echo.Context) error {
	p, err := h.getIdentityProvider(c)
	if err != nil {
		return err
	}

	authRes, err := p.Callback(c.Request().Context(), c.Request())
	if err != nil {
		return c.JSON(http.StatusUnauthorized, err.Error())
	}

	// Gera token JWT do CRM
	claims := jwt.MapClaims{
		"tenant_id": authRes.TenantID,
		"user_id":   authRes.UserID,
		"email":     authRes.Email,
		"iat":       time.Now().Unix(),
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(h.Config.JWTSecret))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "failed to generate token",
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": signed,
	})
}
