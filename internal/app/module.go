package app

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	emiddleware "github.com/labstack/echo/v4/middleware"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/jeanmolossi/verbose-adventure/internal/auth"
	"github.com/jeanmolossi/verbose-adventure/internal/config"
	"github.com/jeanmolossi/verbose-adventure/internal/db"
	"github.com/jeanmolossi/verbose-adventure/internal/http/handlers"
	"github.com/jeanmolossi/verbose-adventure/internal/http/middleware"
	"github.com/jeanmolossi/verbose-adventure/internal/logger"
	"github.com/jeanmolossi/verbose-adventure/internal/repo"
)

// Register all providers and invokes in one place:
func New() *fx.App {
	return fx.New(
		// 1) Builds
		fx.Provide(
			config.New,    // *config.Config
			logger.NewZap, // *zap.Logger

			db.NewMySQL, // *sql.DB (MySQL)

			repo.NewIdentityProviderRepository, // IdentiityProviderRepository

			auth.LoadIdentityProviders, // LoadIdentityProviders interface
			echo.New,                   // *echo.Echo
			handlers.NewAuthHandler,    // *handlers.AuthHandler
			handlers.NewIDPHandler,     // *handlers.IDPHandler

			fx.Annotate(
				middleware.ZapLogger, // func(*zap.Logger) echo.MiddlewareFunc
				fx.ResultTags(`name:"zapMw"`),
			),

			func(cfg *config.Config) middleware.JWTConfig {
				return middleware.JWTConfig{JWTSecret: cfg.JWTSecret}
			},

			fx.Annotate(
				func(cfg *config.Config) echo.MiddlewareFunc {
					return middleware.NewJWTMiddleware(middleware.JWTConfig{
						JWTSecret: cfg.JWTSecret,
					})
				}, fx.ResultTags(`name:"jwtMw"`),
			),
		),
		// 2) Registrations
		fx.Invoke(
			db.RunMigrations,
			registerMiddlewares(),
			registerRoutes(),
			startServer,
		),
	)
}

func registerMiddlewares() any {
	return fx.Annotate(
		func(e *echo.Echo, zl echo.MiddlewareFunc) {
			e.HideBanner = true
			e.Use(emiddleware.Recover())
			e.Use(zl)
		},
		fx.ParamTags(``, `name:"zapMw"`),
	)
}

func startServer(lc fx.Lifecycle, e *echo.Echo, log *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := e.Start(":8081"); err != nil && err != http.ErrServerClosed {
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
