package app

import (
	"database/sql"

	"github.com/jeanmolossi/verbose-adventure/internal/http/handlers"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

func registerRoutes() any {
	return fx.Annotate(
		func(e *echo.Echo, hh *handlers.AuthHandler, jwtMw echo.MiddlewareFunc, mysqlDB *sql.DB) {
			api := e.Group("/api")

			e.GET("/healthz", handlers.Healthz(mysqlDB))

			e.GET("/:tenantID/:idpType/login", hh.Login)
			e.GET("/:tenantID/:idpType/callback", hh.Callback)

			v1 := api.Group("/v1", jwtMw)

			v1.GET("/me", handlers.Me)
		},
		fx.ParamTags(``, ``, `name:"jwtMw"`, ``),
	)
}
