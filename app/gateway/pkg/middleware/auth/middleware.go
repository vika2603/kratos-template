package auth

import (
\t"context"
\t"strings"

\t"github.com/cloudwego/hertz/pkg/app"
\t"github.com/cloudwego/hertz/pkg/protocol/consts"

\tpkgauth "kratos-template/pkg/auth"
)

func New(jwtManager *pkgauth.JWTManager) app.HandlerFunc {
\treturn func(ctx context.Context, c *app.RequestContext) {
\t\tif jwtManager == nil {
\t\t\tc.AbortWithStatusJSON(consts.StatusInternalServerError, map[string]any{
\t\t\t\t"code":    500,
\t\t\t\t"message": "auth middleware not configured",
\t\t\t})
\t\t\treturn
\t\t}

\t\ttokenString, ok := parseBearerToken(string(c.GetHeader("Authorization")))
\t\tif !ok {
\t\t\tc.AbortWithStatusJSON(consts.StatusUnauthorized, map[string]any{
\t\t\t\t"code":    401,
\t\t\t\t"message": "missing or invalid authorization header",
\t\t\t})
\t\t\treturn
\t\t}

\t\tclaims, err := jwtManager.ParseToken(tokenString)
\t\tif err != nil {
\t\t\tc.AbortWithStatusJSON(consts.StatusUnauthorized, map[string]any{
\t\t\t\t"code":    401,
\t\t\t\t"message": "invalid or expired token",
\t\t\t})
\t\t\treturn
\t\t}

\t\tc.Set("claims", claims)
\t\tc.Set("user_id", claims.UserID)
\t\tc.Set("username", claims.Username)
\t\tc.Next(ctx)
\t}
}

func parseBearerToken(value string) (string, bool) {
\tif value == "" {
\t\treturn "", false
\t}

\tparts := strings.SplitN(value, " ", 2)
\tif len(parts) != 2 || parts[0] != "Bearer" {
\t\treturn "", false
\t}

\treturn parts[1], true
}
