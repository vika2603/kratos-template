package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "0123456789abcdef0123456789abcdef"

func newManager(t *testing.T, accessTTL, refreshTTL time.Duration) *JWTManager {
	t.Helper()
	m, err := NewJWTManager(testSecret, accessTTL, refreshTTL)
	if err != nil {
		t.Fatalf("NewJWTManager: %v", err)
	}
	return m
}

func TestNewJWTManagerShortSecret(t *testing.T) {
	if _, err := NewJWTManager("too-short", time.Minute, time.Minute); err == nil {
		t.Fatal("expected error for secret shorter than 32 bytes")
	}
}

func TestTokenRoundtrip(t *testing.T) {
	m := newManager(t, 15*time.Minute, time.Hour)

	tests := []struct {
		name        string
		generate    func() (Token, error)
		wantType    string
		wantUserID  string
		wantSubject string
	}{
		{"access", func() (Token, error) { return m.GenerateAccessToken("u1", "alice") }, TokenTypeAccess, "u1", "u1"},
		{"refresh", func() (Token, error) { return m.GenerateRefreshToken("u1", "alice") }, TokenTypeRefresh, "u1", "u1"},
		{"service", func() (Token, error) { return m.GenerateServiceToken("auth") }, TokenTypeService, "", "auth"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := tt.generate()
			if err != nil {
				t.Fatalf("generate: %v", err)
			}
			if token.JTI == "" {
				t.Error("empty JTI")
			}
			claims, err := m.ParseToken(token.Value, tt.wantType)
			if err != nil {
				t.Fatalf("ParseToken: %v", err)
			}
			if claims.TokenType != tt.wantType {
				t.Errorf("TokenType = %q, want %q", claims.TokenType, tt.wantType)
			}
			if claims.UserID != tt.wantUserID {
				t.Errorf("UserID = %q, want %q", claims.UserID, tt.wantUserID)
			}
			if claims.Subject != tt.wantSubject {
				t.Errorf("Subject = %q, want %q", claims.Subject, tt.wantSubject)
			}
			if claims.Issuer != Issuer {
				t.Errorf("Issuer = %q, want %q", claims.Issuer, Issuer)
			}
			if claims.ID != token.JTI {
				t.Errorf("claims.ID = %q, want %q", claims.ID, token.JTI)
			}
		})
	}
}

func TestParseTokenWrongType(t *testing.T) {
	m := newManager(t, 15*time.Minute, time.Hour)
	token, err := m.GenerateAccessToken("u1", "alice")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if _, err := m.ParseToken(token.Value, TokenTypeRefresh); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("err = %v, want ErrInvalidToken", err)
	}
}

func TestParseTokenWrongSecret(t *testing.T) {
	m := newManager(t, 15*time.Minute, time.Hour)
	other, err := NewJWTManager("ffffffffffffffffffffffffffffffff", 15*time.Minute, time.Hour)
	if err != nil {
		t.Fatalf("NewJWTManager: %v", err)
	}
	token, err := m.GenerateAccessToken("u1", "alice")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if _, err := other.ParseToken(token.Value, ""); !errors.Is(err, ErrInvalidSignature) {
		t.Errorf("err = %v, want ErrInvalidSignature", err)
	}
}

func TestParseTokenExpired(t *testing.T) {
	m := newManager(t, -2*time.Minute, -2*time.Minute)
	token, err := m.GenerateAccessToken("u1", "alice")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if _, err := m.ParseToken(token.Value, ""); !errors.Is(err, ErrExpiredToken) {
		t.Errorf("err = %v, want ErrExpiredToken", err)
	}
}

func TestParseTokenWithinLeeway(t *testing.T) {
	// Expired 10s ago is still valid under the 30s leeway.
	m := newManager(t, -10*time.Second, -10*time.Second)
	token, err := m.GenerateAccessToken("u1", "alice")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if _, err := m.ParseToken(token.Value, TokenTypeAccess); err != nil {
		t.Errorf("ParseToken within leeway: %v", err)
	}
}

func TestParseTokenGarbage(t *testing.T) {
	m := newManager(t, 15*time.Minute, time.Hour)
	if _, err := m.ParseToken("not-a-jwt", ""); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("err = %v, want ErrInvalidToken", err)
	}
}

func TestParseTokenWrongIssuer(t *testing.T) {
	m := newManager(t, 15*time.Minute, time.Hour)
	claims := &Claims{
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "someone-else",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if _, err := m.ParseToken(signed, ""); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("err = %v, want ErrInvalidToken", err)
	}
}

func TestParseTokenNoneAlg(t *testing.T) {
	m := newManager(t, 15*time.Minute, time.Hour)
	claims := &Claims{
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    Issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	// jwt v5 wraps alg-rejection errors inconsistently; only pin what matters.
	if _, err := m.ParseToken(signed, ""); err == nil || errors.Is(err, ErrExpiredToken) {
		t.Errorf("err = %v, want a non-expiry parse error", err)
	}
}
