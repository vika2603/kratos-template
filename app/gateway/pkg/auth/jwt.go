package auth

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/golang-jwt/jwt/v4"
)

var (
	jwtSecret []byte
	once      sync.Once
)

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type ctxKey struct{}

func Init(secret string) {
	once.Do(func() {
		jwtSecret = []byte(secret)
	})
}

func Middleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		authHeader := string(c.GetHeader("Authorization"))
		if authHeader == "" {
			c.AbortWithStatusJSON(consts.StatusUnauthorized, map[string]any{
				"code":    401,
				"message": "missing authorization header",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(consts.StatusUnauthorized, map[string]any{
				"code":    401,
				"message": "invalid authorization header format",
			})
			return
		}

		claims, err := ParseToken(parts[1])
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

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

func GenerateToken(userID int64, username string, expireHours int) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func GetClaims(c *app.RequestContext) *Claims {
	if v, exists := c.Get("claims"); exists {
		if claims, ok := v.(*Claims); ok {
			return claims
		}
	}
	return nil
}

func GetUserID(c *app.RequestContext) int64 {
	if v, exists := c.Get("user_id"); exists {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}
