package app

import (
	"database/sql"

	"github.com/jeanmolossi/verbose-adventure/internal/http/handlers"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

func registerRoutes() any {
	return fx.Annotate(
		func(
			e *echo.Echo,
			hh *handlers.AuthHandler,
			jwtMw echo.MiddlewareFunc,
			idph *handlers.IDPHandler,
			mysqlDB *sql.DB,
		) {
			// Health
			e.GET("/healthz", handlers.Healthz(mysqlDB))

			// Auth SSO
			e.GET("/:tenantID/:idpType/login", hh.Login)
			e.GET("/:tenantID/:idpType/callback", hh.Callback)

			// Admin: Identity providers
			admin := e.Group("/admin/tenants/:tenantID/idps", jwtMw)
			{
				admin.GET(":id", idph.Get)
				admin.GET("", idph.List)
				admin.POST("", idph.Create)
				admin.PUT(":id", idph.Update)
				admin.DELETE(":id", idph.Delete)
			}

			// Protected API
			api := e.Group("/api")
			v1 := api.Group("/v1", jwtMw)
			v1.GET("/me", handlers.Me)
		},
		fx.ParamTags(
			``,             // echo
			``,             // AuthHandler
			`name:"jwtMw"`, // jwt Middleware
			``,             // IDPHandler
			``,             // mysqlDB
		),
	)
}
