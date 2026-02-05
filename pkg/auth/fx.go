package auth

import "go.uber.org/fx"

type Config struct {
	Secret      string
	ExpiryHours int
}

func NewJWTManagerFromConfig(cfg Config) *JWTManager {
	return NewJWTManager(cfg.Secret, cfg.ExpiryHours)
}

var Module = fx.Module("pkg.auth",
	fx.Provide(NewJWTManagerFromConfig),
)
