package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidSignature = errors.New("invalid token signature")
)

// Claims represents JWT claims with user information.
type Claims struct {
	jwt.RegisteredClaims
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// JWTManager handles JWT token operations.
type JWTManager struct {
	secret      []byte
	expiryHours int
}

// NewJWTManager creates a new JWT manager with the given secret and expiry duration.
func NewJWTManager(secret string, expiryHours int) *JWTManager {
	return &JWTManager{
		secret:      []byte(secret),
		expiryHours: expiryHours,
	}
}

// GenerateToken creates a new JWT token for the given user.
func (m *JWTManager) GenerateToken(userID string, username string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(m.expiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// GenerateTokenWithExpiry creates a token with a custom expiry duration in seconds.
func (m *JWTManager) GenerateTokenWithExpiry(userID string, username string, expirySeconds int64) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expirySeconds) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ParseToken parses and validates a JWT token string.
func (m *JWTManager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSignature
		}
		return m.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// ValidateToken checks if a token is valid and returns the claims.
func (m *JWTManager) ValidateToken(tokenString string) (bool, *Claims, error) {
	claims, err := m.ParseToken(tokenString)
	if err != nil {
		return false, nil, err
	}
	return true, claims, nil
}

// ExpiryHours returns the configured token expiry duration in hours.
func (m *JWTManager) ExpiryHours() int {
	return m.expiryHours
}

// ExpirySeconds returns the configured token expiry duration in seconds.
func (m *JWTManager) ExpirySeconds() int64 {
	return int64(m.expiryHours) * 3600
}
