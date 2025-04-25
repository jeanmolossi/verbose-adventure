package handlers

import (
	"database/sql"

	"github.com/labstack/echo/v4"
)

func Healthz(mysqlDB *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := mysqlDB.Ping(); err != nil {
			return c.JSON(500, echo.Map{
				"error": "mysql down",
			})
		}

		return c.JSON(200, echo.Map{
			"mysql": map[string]any{
				"status": "up",
			},
		})
	}
}
