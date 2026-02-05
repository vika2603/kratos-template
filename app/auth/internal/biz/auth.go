package biz

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	pkgauth "kratos-template/pkg/auth"
)

type AuthUserRepo interface {
	GetByUsername(ctx context.Context, username string) (*AuthUser, error)
	GetByID(ctx context.Context, id uint) (*AuthUser, error)
}

type AuthUser struct {
	ID           uint
	Username     string
	PasswordHash string
}

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthUseCase struct {
	repo       AuthUserRepo
	jwtManager *pkgauth.JWTManager
}

func (uc *AuthUseCase) Login(ctx context.Context, username, password string) (string, int64, error) {
	user, err := uc.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return "", 0, ErrInvalidCredentials
		}
		return "", 0, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", 0, ErrInvalidCredentials
	}

	token, err := uc.jwtManager.GenerateToken(int64(user.ID), user.Username)
	if err != nil {
		return "", 0, err
	}

	return token, uc.jwtManager.ExpirySeconds(), nil
}

func (uc *AuthUseCase) Refresh(ctx context.Context, token string) (string, int64, error) {
	claims, err := uc.jwtManager.ParseToken(token)
	if err != nil {
		return "", 0, errors.New("invalid token")
	}

	newToken, err := uc.jwtManager.GenerateToken(claims.UserID, claims.Username)
	if err != nil {
		return "", 0, err
	}

	return newToken, uc.jwtManager.ExpirySeconds(), nil
}

func (uc *AuthUseCase) Validate(ctx context.Context, token string) (bool, uint, string, error) {
	claims, err := uc.jwtManager.ParseToken(token)
	if err != nil {
		return false, 0, "", nil
	}

	user, err := uc.repo.GetByID(ctx, uint(claims.UserID))
	if err != nil {
		return false, 0, "", nil
	}

	return true, user.ID, user.Username, nil
}

func (uc *AuthUseCase) Logout(ctx context.Context, token string) error {
	return nil
}
