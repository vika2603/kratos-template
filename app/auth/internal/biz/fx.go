package biz

import (
	"cmp"
	"os"
	"time"

	"go.uber.org/fx"

	"kratos-template/app/auth/internal/conf"
	pkgauth "kratos-template/pkg/auth"
)

func NewAuthUseCase(repo AuthUserRepo, cfg *conf.Bootstrap) *AuthUseCase {
	expiry := time.Duration(cfg.Auth.TokenExpiry) * time.Second
	if expiry <= 0 {
		expiry = 24 * time.Hour
	}
	secret := cmp.Or(os.Getenv("JWT_SECRET"), cfg.Auth.JwtSecret)
	return &AuthUseCase{
		repo:       repo,
		jwtManager: pkgauth.NewJWTManager(secret, expiry),
	}
}

var Module = fx.Module("auth.biz",
	fx.Provide(NewAuthUseCase),
)
