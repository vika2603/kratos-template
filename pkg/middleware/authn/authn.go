package authn

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	kratosErrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	pkgauth "kratos-template/pkg/auth"
)

const authorizationKey = "authorization"

type claimsKey struct{}

func NewContext(ctx context.Context, claims *pkgauth.Claims) context.Context {
	return context.WithValue(ctx, claimsKey{}, claims)
}

func FromContext(ctx context.Context) (*pkgauth.Claims, bool) {
	claims, ok := ctx.Value(claimsKey{}).(*pkgauth.Claims)
	return claims, ok
}

func Server(manager *pkgauth.JWTManager, allowedTypes ...string) middleware.Middleware {
	allowed := make(map[string]struct{}, len(allowedTypes))
	for _, typ := range allowedTypes {
		allowed[typ] = struct{}{}
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			token, ok := bearerToken(ctx)
			if !ok {
				return nil, kratosErrors.Unauthorized("TOKEN_INVALID", "missing bearer token")
			}
			claims, err := manager.ParseToken(token, "")
			if err != nil {
				return nil, tokenError(err)
			}
			if len(allowed) > 0 {
				if _, ok := allowed[claims.TokenType]; !ok {
					return nil, kratosErrors.Forbidden("TOKEN_INVALID", "token type not allowed")
				}
			}
			return handler(NewContext(ctx, claims), req)
		}
	}
}

func ClientServiceToken(manager *pkgauth.JWTManager, serviceName string) middleware.Middleware {
	var mu sync.Mutex
	var cached pkgauth.Token

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			token, err := serviceToken(manager, serviceName, &mu, &cached)
			if err != nil {
				return nil, kratosErrors.InternalServer("TOKEN_INVALID", "failed to generate service token")
			}
			if tr, ok := transport.FromClientContext(ctx); ok {
				tr.RequestHeader().Set(authorizationKey, "Bearer "+token.Value)
			}
			return handler(ctx, req)
		}
	}
}

func bearerToken(ctx context.Context) (string, bool) {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return "", false
	}
	raw := tr.RequestHeader().Get(authorizationKey)
	if raw == "" {
		raw = tr.RequestHeader().Get("Authorization")
	}
	parts := strings.Fields(raw)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func serviceToken(manager *pkgauth.JWTManager, serviceName string, mu *sync.Mutex, cached *pkgauth.Token) (pkgauth.Token, error) {
	mu.Lock()
	defer mu.Unlock()

	if cached.Value != "" && time.Until(cached.ExpiresAt) > time.Minute {
		return *cached, nil
	}
	token, err := manager.GenerateServiceToken(serviceName)
	if err != nil {
		return pkgauth.Token{}, err
	}
	*cached = token
	return token, nil
}

func tokenError(err error) error {
	if errors.Is(err, pkgauth.ErrExpiredToken) {
		return kratosErrors.Unauthorized("TOKEN_EXPIRED", "token expired")
	}
	return kratosErrors.Unauthorized("TOKEN_INVALID", "invalid token")
}
