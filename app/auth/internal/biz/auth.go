package biz

import (
	"context"
	"errors"
	"kratos-template/pkg/log"
	"time"

	"go.uber.org/zap"

	authv1 "kratos-template/api/auth/v1"
	pkgauth "kratos-template/pkg/auth"
)

type AuthUserRepo interface {
	VerifyCredentials(ctx context.Context, username, password string) (*AuthUser, error)
	GetByID(ctx context.Context, id string) (*AuthUser, error)
}

type TokenRepo interface {
	RevokeAccess(ctx context.Context, jti string, ttl time.Duration) error
	IsAccessRevoked(ctx context.Context, jti string) (bool, error)
	SaveRefresh(ctx context.Context, jti, userID string, ttl time.Duration) error
	ConsumeRefresh(ctx context.Context, jti string) (userID string, ok bool, err error)
	RevokeAllRefresh(ctx context.Context, userID string) error
}

type AuthUser struct {
	ID       string
	Username string
}

type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	TokenType        string
	ExpiresIn        int64
	RefreshExpiresIn int64
}

type AuthUseCase struct {
	userRepo   AuthUserRepo
	tokenRepo  TokenRepo
	jwtManager *pkgauth.JWTManager
	logger     *zap.Logger
}

func (uc *AuthUseCase) Login(ctx context.Context, username, password string) (*TokenPair, error) {
	user, err := uc.userRepo.VerifyCredentials(ctx, username, password)
	if err != nil {
		return nil, err
	}

	pair, err := uc.issueTokenPair(ctx, user)
	if err != nil {
		return nil, uc.internalError(ctx, "failed to issue token", err)
	}
	return pair, nil
}

func (uc *AuthUseCase) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := uc.jwtManager.ParseToken(refreshToken, pkgauth.TokenTypeRefresh)
	if err != nil {
		return nil, authTokenError(err)
	}

	userID, ok, err := uc.tokenRepo.ConsumeRefresh(ctx, claims.ID)
	if err != nil {
		return nil, uc.internalError(ctx, "failed to consume refresh token", err)
	}
	if !ok {
		if err := uc.tokenRepo.RevokeAllRefresh(ctx, claims.UserID); err != nil {
			log.WithContextLogger(ctx, uc.logger).Error("failed to revoke refresh tokens after reuse", zap.Error(err))
		}
		return nil, authv1.ErrorTokenRevoked("refresh token revoked")
	}
	if userID != claims.UserID {
		if err := uc.tokenRepo.RevokeAllRefresh(ctx, claims.UserID); err != nil {
			log.WithContextLogger(ctx, uc.logger).Error("failed to revoke refresh tokens after subject mismatch", zap.Error(err))
		}
		return nil, authv1.ErrorTokenRevoked("refresh token revoked")
	}

	user, err := uc.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	pair, err := uc.issueTokenPair(ctx, user)
	if err != nil {
		return nil, uc.internalError(ctx, "failed to issue token", err)
	}
	return pair, nil
}

func (uc *AuthUseCase) Validate(ctx context.Context, token string) (bool, string, string, error) {
	claims, err := uc.jwtManager.ParseToken(token, pkgauth.TokenTypeAccess)
	if err != nil {
		return false, "", "", authTokenError(err)
	}

	revoked, err := uc.tokenRepo.IsAccessRevoked(ctx, claims.ID)
	if err != nil {
		return false, "", "", uc.internalError(ctx, "failed to validate token", err)
	}
	if revoked {
		return false, "", "", authv1.ErrorTokenRevoked("token revoked")
	}

	// token can outlive its user (deleted/disabled), so re-check
	user, err := uc.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return false, "", "", err
	}

	return true, user.ID, user.Username, nil
}

func (uc *AuthUseCase) Logout(ctx context.Context, accessToken, refreshToken string) error {
	claims, err := uc.jwtManager.ParseToken(accessToken, pkgauth.TokenTypeAccess)
	if err == nil {
		if ttl := time.Until(claims.ExpiresAt.Time); ttl > 0 {
			if err := uc.tokenRepo.RevokeAccess(ctx, claims.ID, ttl); err != nil {
				return uc.internalError(ctx, "failed to revoke access token", err)
			}
		}
	}

	if refreshToken != "" {
		refreshClaims, err := uc.jwtManager.ParseToken(refreshToken, pkgauth.TokenTypeRefresh)
		if err == nil {
			if _, _, err := uc.tokenRepo.ConsumeRefresh(ctx, refreshClaims.ID); err != nil {
				return uc.internalError(ctx, "failed to revoke refresh token", err)
			}
		}
	}
	return nil
}

func (uc *AuthUseCase) issueTokenPair(ctx context.Context, user *AuthUser) (*TokenPair, error) {
	access, err := uc.jwtManager.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}
	refresh, err := uc.jwtManager.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}
	if err := uc.tokenRepo.SaveRefresh(ctx, refresh.JTI, user.ID, time.Until(refresh.ExpiresAt)); err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:      access.Value,
		RefreshToken:     refresh.Value,
		TokenType:        "Bearer",
		ExpiresIn:        uc.jwtManager.AccessExpirySeconds(),
		RefreshExpiresIn: uc.jwtManager.RefreshExpirySeconds(),
	}, nil
}

func authTokenError(err error) error {
	switch {
	case errors.Is(err, pkgauth.ErrExpiredToken):
		return authv1.ErrorTokenExpired("token expired")
	case errors.Is(err, pkgauth.ErrInvalidSignature), errors.Is(err, pkgauth.ErrInvalidToken):
		return authv1.ErrorTokenInvalid("invalid token")
	default:
		return authv1.ErrorTokenInvalid("invalid token")
	}
}

func (uc *AuthUseCase) internalError(ctx context.Context, message string, err error) error {
	log.WithContextLogger(ctx, uc.logger).Error(message, zap.Error(err))
	return authv1.ErrorInternal("%s", message)
}
