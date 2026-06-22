package biz

import (
	"context"

	kratosErrors "github.com/go-kratos/kratos/v2/errors"

	pkgauth "kratos-template/pkg/auth"
)

type AuthUserRepo interface {
	VerifyCredentials(ctx context.Context, username, password string) (*AuthUser, error)
	GetByID(ctx context.Context, id string) (*AuthUser, error)
}

type AuthUser struct {
	ID       string
	Username string
}

var ErrInvalidCredentials = kratosErrors.Unauthorized("INVALID_CREDENTIALS", "invalid credentials")

type AuthUseCase struct {
	repo       AuthUserRepo
	jwtManager *pkgauth.JWTManager
}

func (uc *AuthUseCase) Login(ctx context.Context, username, password string) (string, int64, error) {
	user, err := uc.repo.VerifyCredentials(ctx, username, password)
	if err != nil {
		return "", 0, err
	}

	token, err := uc.jwtManager.GenerateToken(user.ID, user.Username)
	if err != nil {
		return "", 0, err
	}

	return token, uc.jwtManager.ExpirySeconds(), nil
}

func (uc *AuthUseCase) Refresh(ctx context.Context, token string) (string, int64, error) {
	claims, err := uc.jwtManager.ParseToken(token)
	if err != nil {
		return "", 0, ErrInvalidCredentials
	}

	newToken, err := uc.jwtManager.GenerateToken(claims.UserID, claims.Username)
	if err != nil {
		return "", 0, err
	}

	return newToken, uc.jwtManager.ExpirySeconds(), nil
}

func (uc *AuthUseCase) Validate(ctx context.Context, token string) (bool, string, string, error) {
	claims, err := uc.jwtManager.ParseToken(token)
	if err != nil {
		return false, "", "", ErrInvalidCredentials
	}

	// token can outlive its user (deleted/disabled), so re-check
	user, err := uc.repo.GetByID(ctx, claims.UserID)
	if err != nil {
		return false, "", "", err
	}

	return true, user.ID, user.Username, nil
}

func (uc *AuthUseCase) Logout(ctx context.Context, token string) error {
	// no-op: stateless JWT. TODO: token denylist (e.g. Redis) to truly revoke.
	return nil
}
