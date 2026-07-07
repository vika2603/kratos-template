package biz

import (
	"cmp"
	"kratos-template/app/auth/internal/conf"
	"os"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	pkgauth "kratos-template/pkg/auth"
)

func NewAuthUseCase(userRepo AuthUserRepo, tokenRepo TokenRepo, loginGuard LoginGuardRepo, cfg *conf.Bootstrap, logger *zap.Logger) (*AuthUseCase, error) {
	accessTTL := time.Duration(cfg.Auth.AccessTokenExpiry) * time.Second
	if accessTTL <= 0 {
		accessTTL = 15 * time.Minute
	}
	refreshTTL := time.Duration(cfg.Auth.RefreshTokenExpiry) * time.Second
	if refreshTTL <= 0 {
		refreshTTL = 168 * time.Hour
	}
	secret := cmp.Or(os.Getenv("JWT_SECRET"), cfg.Auth.JwtSecret)
	manager, err := pkgauth.NewJWTManager(secret, accessTTL, refreshTTL)
	if err != nil {
		return nil, err
	}
	return &AuthUseCase{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		loginGuard: loginGuard,
		jwtManager: manager,
		logger:     logger,
	}, nil
}

var Module = fx.Module("auth.biz",
	fx.Provide(NewAuthUseCase),
)
