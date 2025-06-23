package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConfigOption func(*pgxpool.Config)

func WithSetLocal(hook func(context.Context, *pgx.Conn) bool) ConfigOption {
	return func(c *pgxpool.Config) {
		original := c.BeforeAcquire

		c.BeforeAcquire = func(ctx context.Context, c *pgx.Conn) bool {
			if !hook(ctx, c) {
				return false
			}

			return original(ctx, c)
		}
	}
}
