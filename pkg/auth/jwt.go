package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidSignature = errors.New("invalid token signature")
)

const (
	Issuer           = "kratos-template"
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
	TokenTypeService = "service"

	minSecretBytes = 32
	serviceTTL     = 5 * time.Minute
)

type Token struct {
	Value     string
	JTI       string
	ExpiresAt time.Time
}

type Claims struct {
	jwt.RegisteredClaims
	UserID    string `json:"user_id,omitempty"`
	Username  string `json:"username,omitempty"`
	TokenType string `json:"token_type"`
}

type JWTManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWTManager(secret string, accessTTL, refreshTTL time.Duration) (*JWTManager, error) {
	if len([]byte(secret)) < minSecretBytes {
		return nil, fmt.Errorf("jwt secret must be at least %d bytes", minSecretBytes)
	}
	return &JWTManager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}, nil
}

func (m *JWTManager) GenerateAccessToken(userID string, username string) (Token, error) {
	return m.generateToken(userID, username, TokenTypeAccess, m.accessTTL, userID)
}

func (m *JWTManager) GenerateRefreshToken(userID string, username string) (Token, error) {
	return m.generateToken(userID, username, TokenTypeRefresh, m.refreshTTL, userID)
}

func (m *JWTManager) GenerateServiceToken(serviceName string) (Token, error) {
	return m.generateToken("", "", TokenTypeService, serviceTTL, serviceName)
}

func (m *JWTManager) ParseToken(tokenString string, wantType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidSignature
			}
			return m.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithLeeway(30*time.Second),
		jwt.WithIssuer(Issuer),
	)
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, ErrExpiredToken
		case errors.Is(err, jwt.ErrTokenSignatureInvalid), errors.Is(err, ErrInvalidSignature):
			return nil, ErrInvalidSignature
		default:
			return nil, ErrInvalidToken
		}
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	if wantType != "" && claims.TokenType != wantType {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (m *JWTManager) AccessExpirySeconds() int64 {
	return int64(m.accessTTL.Seconds())
}

func (m *JWTManager) RefreshExpirySeconds() int64 {
	return int64(m.refreshTTL.Seconds())
}

func (m *JWTManager) generateToken(userID, username, tokenType string, ttl time.Duration, subject string) (Token, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)
	jti := uuid.NewString()
	claims := &Claims{
		UserID:    userID,
		Username:  username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   subject,
			Issuer:    Issuer,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	value, err := token.SignedString(m.secret)
	if err != nil {
		return Token{}, err
	}
	return Token{
		Value:     value,
		JTI:       jti,
		ExpiresAt: expiresAt,
	}, nil
}
