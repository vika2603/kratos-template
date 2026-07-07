package authn

import (
	"context"
	"testing"
	"time"

	kratosErrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/transport"

	pkgauth "kratos-template/pkg/auth"
)

const testSecret = "0123456789abcdef0123456789abcdef"

type mapHeader map[string]string

func (h mapHeader) Get(key string) string      { return h[key] }
func (h mapHeader) Set(key, value string)      { h[key] = value }
func (h mapHeader) Add(key, value string)      { h[key] = value }
func (h mapHeader) Keys() []string             { return nil }
func (h mapHeader) Values(key string) []string { return []string{h[key]} }

type fakeTransport struct {
	header mapHeader
}

func (f *fakeTransport) Kind() transport.Kind            { return transport.KindGRPC }
func (f *fakeTransport) Endpoint() string                { return "" }
func (f *fakeTransport) Operation() string               { return "/test.Service/Method" }
func (f *fakeTransport) RequestHeader() transport.Header { return f.header }
func (f *fakeTransport) ReplyHeader() transport.Header   { return f.header }

func serverCtx(authorization string) context.Context {
	header := mapHeader{}
	if authorization != "" {
		header.Set("authorization", authorization)
	}
	return transport.NewServerContext(context.Background(), &fakeTransport{header: header})
}

func newManager(t *testing.T) *pkgauth.JWTManager {
	t.Helper()
	m, err := pkgauth.NewJWTManager(testSecret, 15*time.Minute, time.Hour)
	if err != nil {
		t.Fatalf("NewJWTManager: %v", err)
	}
	return m
}

func passthrough(ctx context.Context, req any) (any, error) { return req, nil }

func TestServerMissingOrMalformedToken(t *testing.T) {
	m := newManager(t)
	mw := Server(m, pkgauth.TokenTypeAccess)(passthrough)

	tests := []struct {
		name string
		ctx  context.Context
	}{
		{"no transport", context.Background()},
		{"no header", serverCtx("")},
		{"not bearer", serverCtx("Basic dXNlcjpwYXNz")},
		{"bearer without token", serverCtx("Bearer")},
		{"garbage token", serverCtx("Bearer not-a-jwt")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mw(tt.ctx, nil)
			if !kratosErrors.IsUnauthorized(err) {
				t.Errorf("err = %v, want Unauthorized", err)
			}
		})
	}
}

func TestServerAllowedTypeInjectsClaims(t *testing.T) {
	m := newManager(t)
	token, err := m.GenerateAccessToken("u1", "alice")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	var got *pkgauth.Claims
	handler := func(ctx context.Context, req any) (any, error) {
		got, _ = FromContext(ctx)
		return req, nil
	}
	if _, err := Server(m, pkgauth.TokenTypeAccess)(handler)(serverCtx("Bearer "+token.Value), nil); err != nil {
		t.Fatalf("middleware: %v", err)
	}
	if got == nil || got.UserID != "u1" {
		t.Errorf("claims = %+v, want UserID u1", got)
	}
}

func TestServerDisallowedTypeForbidden(t *testing.T) {
	m := newManager(t)
	token, err := m.GenerateServiceToken("auth")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	mw := Server(m, pkgauth.TokenTypeAccess)(passthrough)

	_, err = mw(serverCtx("Bearer "+token.Value), nil)
	if !kratosErrors.IsForbidden(err) {
		t.Errorf("err = %v, want Forbidden", err)
	}
}

func TestServerExpiredToken(t *testing.T) {
	expired, err := pkgauth.NewJWTManager(testSecret, -2*time.Minute, -2*time.Minute)
	if err != nil {
		t.Fatalf("NewJWTManager: %v", err)
	}
	token, err := expired.GenerateAccessToken("u1", "alice")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	mw := Server(newManager(t))(passthrough)

	_, err = mw(serverCtx("Bearer "+token.Value), nil)
	e := kratosErrors.FromError(err)
	if !kratosErrors.IsUnauthorized(err) || e.Reason != "TOKEN_EXPIRED" {
		t.Errorf("err = %v, want Unauthorized TOKEN_EXPIRED", err)
	}
}
