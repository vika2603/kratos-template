package biz

import (
	"go.uber.org/fx"

	"kratos-template/app/auth/internal/conf"
	pkgauth "kratos-template/pkg/auth"
)

func NewAuthUseCase(repo AuthUserRepo, cfg *conf.Bootstrap) *AuthUseCase {
	expiryHours := int(cfg.Auth.TokenExpiry / 3600)
	if expiryHours < 1 {
		expiryHours = 1
	}
	return &AuthUseCase{
		repo:       repo,
		jwtManager: pkgauth.NewJWTManager(cfg.Auth.JwtSecret, expiryHours),
	}
}

var Module = fx.Module("auth.biz",
	fx.Provide(NewAuthUseCase),
)
