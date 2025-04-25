package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Me(c echo.Context) error {
	tenantID := c.Get("tenant_id")
	userID := c.Get("user_id")
	email := c.Get("email")

	return c.JSON(http.StatusOK, echo.Map{
		"tenant_id": tenantID,
		"user_id":   userID,
		"email":     email,
	})
}
