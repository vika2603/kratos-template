package auth

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	pkgauth "kratos-template/pkg/auth"
)

var defaultMiddleware app.HandlerFunc

func Init(jwtManager *pkgauth.JWTManager) {
	defaultMiddleware = New(jwtManager)
}

func Default() app.HandlerFunc {
	return defaultMiddleware
}

func New(jwtManager *pkgauth.JWTManager) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if jwtManager == nil {
			c.AbortWithStatusJSON(consts.StatusInternalServerError, map[string]any{
				"code":    500,
				"message": "auth middleware not configured",
			})
			return
		}

		tokenString, ok := parseBearerToken(string(c.GetHeader("Authorization")))
		if !ok {
			c.AbortWithStatusJSON(consts.StatusUnauthorized, map[string]any{
				"code":    401,
				"message": "missing or invalid authorization header",
			})
			return
		}

		claims, err := jwtManager.ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(consts.StatusUnauthorized, map[string]any{
				"code":    401,
				"message": "invalid or expired token",
			})
			return
		}

		c.Set("claims", claims)
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next(ctx)
	}
}

func parseBearerToken(value string) (string, bool) {
	if value == "" {
		return "", false
	}

	parts := strings.SplitN(value, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", false
	}

	return parts[1], true
}
