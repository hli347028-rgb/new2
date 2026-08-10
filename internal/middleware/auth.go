package middleware

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
)

type tokenKey struct{}

// BearerToken extracts Authorization: Bearer <token> into request context.
func BearerToken() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			token := extractHTTPToken(ctx)
			if token == "" {
				token = extractGRPCToken(ctx)
			}
			if token != "" {
				ctx = context.WithValue(ctx, tokenKey{}, token)
			}
			return handler(ctx, req)
		}
	}
}

func extractHTTPToken(ctx context.Context) string {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return ""
	}
	ht, ok := tr.(*http.Transport)
	if !ok {
		return ""
	}
	if token := parseBearer(ht.RequestHeader().Get("Authorization")); token != "" {
		return token
	}
	// 兼容直接放在请求头的 token / Access-Token
	if token := strings.TrimSpace(ht.RequestHeader().Get("Access-Token")); token != "" {
		return token
	}
	return strings.TrimSpace(ht.RequestHeader().Get("token"))
}

// ParseBearer extracts token from an Authorization header value.
func ParseBearer(value string) string {
	return parseBearer(value)
}

func extractGRPCToken(ctx context.Context) string {
	md, ok := metadata.FromServerContext(ctx)
	if !ok {
		return ""
	}
	return parseBearer(md.Get("authorization"))
}

func parseBearer(value string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(value, prefix) {
		return ""
	}
	return strings.TrimSpace(value[len(prefix):])
}

// TokenFromContext returns JWT from Authorization header stored in context.
func TokenFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(tokenKey{}).(string); ok {
		return v
	}
	return ""
}

// ResolveToken prefers Authorization header token, falls back to legacy field.
func ResolveToken(ctx context.Context, fallback string) string {
	if t := TokenFromContext(ctx); t != "" {
		return t
	}
	return fallback
}
