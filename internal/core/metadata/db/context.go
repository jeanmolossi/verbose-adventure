package db

import (
	"context"
	"errors"
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
