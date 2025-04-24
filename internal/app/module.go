package app

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	emiddleware "github.com/labstack/echo/v4/middleware"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/jeanmolossi/verbose-adventure/internal/config"
	"github.com/jeanmolossi/verbose-adventure/internal/db"
	"github.com/jeanmolossi/verbose-adventure/internal/http/middleware"
	"github.com/jeanmolossi/verbose-adventure/internal/logger"
)

// Register all providers and invokes in one place:
func New() *fx.App {
	return fx.New(
		// 1) Builds
		fx.Provide(
			config.New,           // *config.Config
			logger.NewZap,        // *zap.Logger
			db.NewMySQL,          // *sql.DB (MySQL) db.NewPostgres,       // *sql.DB (Postgres)
			echo.New,             // *echo.Echo
			middleware.ZapLogger, // func(*zap.Logger) echo.MiddlewareFunc
			// handlers.NewUserHandler, // exemplo de handler
			// … outros handlers/repos/providers
		),
		// 2) Registrations
		fx.Invoke(
			registerDBHealth,
			db.RunMigrations,
			registerMiddlewares,
			registerRoutes,
			startServer,
		),
	)
}

func registerDBHealth(e *echo.Echo, mysqlDB *sql.DB) {
	e.GET("/healthz", func(c echo.Context) error {
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
	})
}

func registerMiddlewares(e *echo.Echo, zl echo.MiddlewareFunc) {
	e.HideBanner = true
	e.Use(emiddleware.Recover())
	e.Use(zl)
}

func registerRoutes(e *echo.Echo) {
	// grp := e.Group("/users")
	// grp.GET("", uh.List)
	// grp.POST("", uh.Create)
	// … outras rotas
}

func startServer(lc fx.Lifecycle, e *echo.Echo, log *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
					log.Fatal("Echo start failed", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Shutting down")
			return e.Shutdown(ctx)
		},
	})
}
