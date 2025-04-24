package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func ZapLogger(l *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			end := time.Since(start)

			l.Info("http request",
				zap.Dict("request",
					zap.String("method", c.Request().Method),
					zap.String("path", c.Request().URL.Path),
					zap.String("query_string", c.QueryString()),
					zap.Any("headers", c.Request().Header),
				),
				zap.Dict("response",
					zap.Int("status", c.Response().Status),
					zap.Duration("latency", end),
					zap.String("human_latency", end.String()),
					zap.Any("headers", c.Response().Header()),
				),
				zap.Error(err),
			)

			return err
		}
	}
}
