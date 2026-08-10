package service

import (
	"context"

	"backend/internal/middleware"
)

func resolveToken(ctx context.Context, fallback string) string {
	return middleware.ResolveToken(ctx, fallback)
}
