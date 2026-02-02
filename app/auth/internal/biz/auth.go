package biz

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
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
	repo        AuthUserRepo
	jwtSecret   string
	tokenExpiry int64
}

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
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

	token, err := uc.generateToken(user.ID, user.Username)
	if err != nil {
		return "", 0, err
	}

	return token, uc.tokenExpiry, nil
}

func (uc *AuthUseCase) Refresh(ctx context.Context, token string) (string, int64, error) {
	claims, err := uc.parseToken(token)
	if err != nil {
		return "", 0, errors.New("invalid token")
	}

	newToken, err := uc.generateToken(claims.UserID, claims.Username)
	if err != nil {
		return "", 0, err
	}

	return newToken, uc.tokenExpiry, nil
}

func (uc *AuthUseCase) Validate(ctx context.Context, token string) (bool, uint, string, error) {
	claims, err := uc.parseToken(token)
	if err != nil {
		return false, 0, "", nil
	}

	user, err := uc.repo.GetByID(ctx, claims.UserID)
	if err != nil {
		return false, 0, "", nil
	}

	return true, user.ID, user.Username, nil
}

func (uc *AuthUseCase) Logout(ctx context.Context, token string) error {
	return nil
}

func (uc *AuthUseCase) generateToken(userID uint, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(uc.tokenExpiry) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.jwtSecret))
}

func (uc *AuthUseCase) parseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(uc.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
