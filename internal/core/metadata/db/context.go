package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type tenantIDKey struct{}

var ErrNoTenantIDinCtx = errors.New("no tenant id on context")

func StartSessionWith(ctx context.Context, tenantID int) context.Context {
	return context.WithValue(ctx, tenantIDKey{}, tenantID)
}

func TenantIDFromCtx(ctx context.Context) (int, error) {
	id, ok := ctx.Value(tenantIDKey{}).(int)
	if !ok {
		return -1, ErrNoTenantIDinCtx
	}

	return id, nil
}

func setTenant(pool *pgxpool.Config) {
	WithSetLocal(func(ctx context.Context, c *pgx.Conn) bool {
		tenantID, err := TenantIDFromCtx(ctx)
		if err != nil {
			return false
		}

		// sets the RLS var into transaction scope
		if _, err := c.Exec(ctx,
			"SET LOCAL app.current_tenant = $1", tenantID); err != nil {
			return false
		}

		return true
	})(pool)
}
