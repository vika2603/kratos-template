package biz

import (
	"go.uber.org/fx"

	"kratos-template/app/auth/internal/conf"
)

func NewAuthUseCase(repo AuthUserRepo, cfg *conf.Bootstrap) *AuthUseCase {
	return &AuthUseCase{
		repo:        repo,
		jwtSecret:   cfg.Auth.JwtSecret,
		tokenExpiry: cfg.Auth.TokenExpiry,
	}
}

var Module = fx.Module("auth.biz",
	fx.Provide(NewAuthUseCase),
)
