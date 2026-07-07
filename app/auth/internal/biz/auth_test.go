package biz

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"

	authv1 "kratos-template/api/auth/v1"
	userv1 "kratos-template/api/user/v1"
	pkgauth "kratos-template/pkg/auth"
)

const testSecret = "0123456789abcdef0123456789abcdef"

type fakeUserRepo struct {
	verify  func(ctx context.Context, username, password string) (*AuthUser, error)
	getByID func(ctx context.Context, id string) (*AuthUser, error)
}

func (f *fakeUserRepo) VerifyCredentials(ctx context.Context, username, password string) (*AuthUser, error) {
	return f.verify(ctx, username, password)
}

func (f *fakeUserRepo) GetByID(ctx context.Context, id string) (*AuthUser, error) {
	return f.getByID(ctx, id)
}

type savedRefresh struct {
	jti    string
	userID string
	ttl    time.Duration
}

type fakeTokenRepo struct {
	refresh map[string]string // jti -> userID
	revoked map[string]time.Duration

	saved          []savedRefresh
	revokeAllUsers []string

	saveErr    error
	consumeErr error
	revokedErr error
}

func newFakeTokenRepo() *fakeTokenRepo {
	return &fakeTokenRepo{
		refresh: make(map[string]string),
		revoked: make(map[string]time.Duration),
	}
}

func (f *fakeTokenRepo) RevokeAccess(_ context.Context, jti string, ttl time.Duration) error {
	f.revoked[jti] = ttl
	return nil
}

func (f *fakeTokenRepo) IsAccessRevoked(_ context.Context, jti string) (bool, error) {
	if f.revokedErr != nil {
		return false, f.revokedErr
	}
	_, ok := f.revoked[jti]
	return ok, nil
}

func (f *fakeTokenRepo) SaveRefresh(_ context.Context, jti, userID string, ttl time.Duration) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.refresh[jti] = userID
	f.saved = append(f.saved, savedRefresh{jti, userID, ttl})
	return nil
}

func (f *fakeTokenRepo) ConsumeRefresh(_ context.Context, jti string) (string, bool, error) {
	if f.consumeErr != nil {
		return "", false, f.consumeErr
	}
	userID, ok := f.refresh[jti]
	if !ok {
		return "", false, nil
	}
	delete(f.refresh, jti)
	return userID, true, nil
}

func (f *fakeTokenRepo) RevokeAllRefresh(_ context.Context, userID string) error {
	f.revokeAllUsers = append(f.revokeAllUsers, userID)
	return nil
}

type fakeLoginGuard struct {
	failures map[string]int64
	err      error
}

func newFakeLoginGuard() *fakeLoginGuard {
	return &fakeLoginGuard{failures: make(map[string]int64)}
}

func (f *fakeLoginGuard) FailureCount(_ context.Context, username string) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	return f.failures[username], nil
}

func (f *fakeLoginGuard) RecordFailure(_ context.Context, username string, _ time.Duration) error {
	f.failures[username]++
	return nil
}

func (f *fakeLoginGuard) Reset(_ context.Context, username string) error {
	delete(f.failures, username)
	return nil
}

func newTestUseCase(t *testing.T, userRepo AuthUserRepo, tokenRepo TokenRepo) *AuthUseCase {
	t.Helper()
	manager, err := pkgauth.NewJWTManager(testSecret, 15*time.Minute, time.Hour)
	if err != nil {
		t.Fatalf("NewJWTManager: %v", err)
	}
	return &AuthUseCase{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		loginGuard: newFakeLoginGuard(),
		jwtManager: manager,
		logger:     zap.NewNop(),
	}
}

func happyUserRepo() *fakeUserRepo {
	user := &AuthUser{ID: "u1", Username: "alice"}
	return &fakeUserRepo{
		verify:  func(context.Context, string, string) (*AuthUser, error) { return user, nil },
		getByID: func(context.Context, string) (*AuthUser, error) { return user, nil },
	}
}

func TestLoginSuccess(t *testing.T) {
	tokens := newFakeTokenRepo()
	uc := newTestUseCase(t, happyUserRepo(), tokens)

	pair, err := uc.Login(context.Background(), "alice", "password123")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if pair.ExpiresIn != 900 || pair.RefreshExpiresIn != 3600 {
		t.Errorf("expiries = %d/%d, want 900/3600", pair.ExpiresIn, pair.RefreshExpiresIn)
	}
	if pair.TokenType != "Bearer" {
		t.Errorf("TokenType = %q", pair.TokenType)
	}
	if len(tokens.saved) != 1 {
		t.Fatalf("SaveRefresh called %d times, want 1", len(tokens.saved))
	}
	saved := tokens.saved[0]
	if saved.userID != "u1" {
		t.Errorf("saved userID = %q, want u1", saved.userID)
	}
	if saved.ttl < 59*time.Minute || saved.ttl > time.Hour {
		t.Errorf("saved ttl = %v, want ~1h", saved.ttl)
	}
	claims, err := uc.jwtManager.ParseToken(pair.RefreshToken, pkgauth.TokenTypeRefresh)
	if err != nil {
		t.Fatalf("parse issued refresh: %v", err)
	}
	if claims.ID != saved.jti {
		t.Errorf("saved jti = %q, want %q from issued token", saved.jti, claims.ID)
	}
}

func TestLoginInvalidCredentialsPassthrough(t *testing.T) {
	repo := &fakeUserRepo{
		verify: func(context.Context, string, string) (*AuthUser, error) {
			return nil, userv1.ErrorInvalidCredentials("invalid credentials")
		},
	}
	uc := newTestUseCase(t, repo, newFakeTokenRepo())

	_, err := uc.Login(context.Background(), "alice", "wrong")
	if !userv1.IsInvalidCredentials(err) {
		t.Errorf("err = %v, want INVALID_CREDENTIALS passthrough", err)
	}
}

func TestLoginBruteForceThrottled(t *testing.T) {
	verifyCalls := 0
	repo := &fakeUserRepo{
		verify: func(context.Context, string, string) (*AuthUser, error) {
			verifyCalls++
			return nil, authv1.ErrorInvalidCredentials("invalid credentials")
		},
	}
	uc := newTestUseCase(t, repo, newFakeTokenRepo())

	for i := range maxLoginFailures {
		if _, err := uc.Login(context.Background(), "alice", "wrong"); !authv1.IsInvalidCredentials(err) {
			t.Fatalf("attempt %d: err = %v, want INVALID_CREDENTIALS", i+1, err)
		}
	}
	_, err := uc.Login(context.Background(), "alice", "wrong")
	if !authv1.IsRateLimited(err) {
		t.Errorf("err = %v, want RATE_LIMITED", err)
	}
	if verifyCalls != maxLoginFailures {
		t.Errorf("VerifyCredentials called %d times, want %d (throttled attempt must not hit user service)",
			verifyCalls, maxLoginFailures)
	}
}

func TestLoginSuccessResetsFailures(t *testing.T) {
	uc := newTestUseCase(t, happyUserRepo(), newFakeTokenRepo())
	guard := uc.loginGuard.(*fakeLoginGuard)
	guard.failures["alice"] = maxLoginFailures - 1

	if _, err := uc.Login(context.Background(), "alice", "password123"); err != nil {
		t.Fatalf("Login: %v", err)
	}
	if guard.failures["alice"] != 0 {
		t.Errorf("failures = %d, want reset to 0", guard.failures["alice"])
	}
}

func TestLoginGuardFailsOpen(t *testing.T) {
	uc := newTestUseCase(t, happyUserRepo(), newFakeTokenRepo())
	uc.loginGuard.(*fakeLoginGuard).err = errors.New("redis down")

	if _, err := uc.Login(context.Background(), "alice", "password123"); err != nil {
		t.Errorf("Login with unavailable guard: %v, want success (fail-open)", err)
	}
}

func TestLoginSaveRefreshFails(t *testing.T) {
	tokens := newFakeTokenRepo()
	tokens.saveErr = errors.New("redis down")
	uc := newTestUseCase(t, happyUserRepo(), tokens)

	_, err := uc.Login(context.Background(), "alice", "password123")
	if !authv1.IsInternal(err) {
		t.Errorf("err = %v, want INTERNAL", err)
	}
}

// login issues a pair and returns the refresh JTI so tests can manipulate state.
func login(t *testing.T, uc *AuthUseCase) (pair *TokenPair, refreshJTI string) {
	t.Helper()
	pair, err := uc.Login(context.Background(), "alice", "password123")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	claims, err := uc.jwtManager.ParseToken(pair.RefreshToken, pkgauth.TokenTypeRefresh)
	if err != nil {
		t.Fatalf("parse refresh: %v", err)
	}
	return pair, claims.ID
}

func TestRefreshRotates(t *testing.T) {
	tokens := newFakeTokenRepo()
	uc := newTestUseCase(t, happyUserRepo(), tokens)
	pair, oldJTI := login(t, uc)

	newPair, err := uc.Refresh(context.Background(), pair.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if _, ok := tokens.refresh[oldJTI]; ok {
		t.Error("old refresh JTI not consumed")
	}
	claims, err := uc.jwtManager.ParseToken(newPair.RefreshToken, pkgauth.TokenTypeRefresh)
	if err != nil {
		t.Fatalf("parse new refresh: %v", err)
	}
	if claims.ID == oldJTI {
		t.Error("refresh token was not rotated")
	}
	if _, ok := tokens.refresh[claims.ID]; !ok {
		t.Error("new refresh JTI not saved")
	}
}

func TestRefreshReuseRevokesAll(t *testing.T) {
	tokens := newFakeTokenRepo()
	uc := newTestUseCase(t, happyUserRepo(), tokens)
	pair, _ := login(t, uc)

	if _, err := uc.Refresh(context.Background(), pair.RefreshToken); err != nil {
		t.Fatalf("first Refresh: %v", err)
	}
	_, err := uc.Refresh(context.Background(), pair.RefreshToken)
	if !authv1.IsTokenRevoked(err) {
		t.Errorf("err = %v, want TOKEN_REVOKED", err)
	}
	if len(tokens.revokeAllUsers) != 1 || tokens.revokeAllUsers[0] != "u1" {
		t.Errorf("revokeAllUsers = %v, want [u1]", tokens.revokeAllUsers)
	}
}

func TestRefreshSubjectMismatchRevokesAll(t *testing.T) {
	tokens := newFakeTokenRepo()
	uc := newTestUseCase(t, happyUserRepo(), tokens)
	pair, jti := login(t, uc)
	tokens.refresh[jti] = "someone-else"

	_, err := uc.Refresh(context.Background(), pair.RefreshToken)
	if !authv1.IsTokenRevoked(err) {
		t.Errorf("err = %v, want TOKEN_REVOKED", err)
	}
	if len(tokens.revokeAllUsers) != 1 || tokens.revokeAllUsers[0] != "u1" {
		t.Errorf("revokeAllUsers = %v, want [u1]", tokens.revokeAllUsers)
	}
}

func TestRefreshExpiredToken(t *testing.T) {
	uc := newTestUseCase(t, happyUserRepo(), newFakeTokenRepo())
	expired, err := pkgauth.NewJWTManager(testSecret, -2*time.Minute, -2*time.Minute)
	if err != nil {
		t.Fatalf("NewJWTManager: %v", err)
	}
	token, err := expired.GenerateRefreshToken("u1", "alice")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if _, err := uc.Refresh(context.Background(), token.Value); !authv1.IsTokenExpired(err) {
		t.Errorf("err = %v, want TOKEN_EXPIRED", err)
	}
}

func TestRefreshRejectsAccessToken(t *testing.T) {
	tokens := newFakeTokenRepo()
	uc := newTestUseCase(t, happyUserRepo(), tokens)
	pair, _ := login(t, uc)

	if _, err := uc.Refresh(context.Background(), pair.AccessToken); !authv1.IsTokenInvalid(err) {
		t.Errorf("err = %v, want TOKEN_INVALID", err)
	}
}

func TestValidate(t *testing.T) {
	tokens := newFakeTokenRepo()
	uc := newTestUseCase(t, happyUserRepo(), tokens)
	pair, _ := login(t, uc)

	valid, userID, username, err := uc.Validate(context.Background(), pair.AccessToken)
	if err != nil || !valid {
		t.Fatalf("Validate = (%v, %v), want valid", valid, err)
	}
	if userID != "u1" || username != "alice" {
		t.Errorf("identity = %q/%q, want u1/alice", userID, username)
	}
}

func TestValidateRevoked(t *testing.T) {
	tokens := newFakeTokenRepo()
	uc := newTestUseCase(t, happyUserRepo(), tokens)
	pair, _ := login(t, uc)

	claims, err := uc.jwtManager.ParseToken(pair.AccessToken, pkgauth.TokenTypeAccess)
	if err != nil {
		t.Fatalf("parse access: %v", err)
	}
	tokens.revoked[claims.ID] = time.Minute

	if _, _, _, err := uc.Validate(context.Background(), pair.AccessToken); !authv1.IsTokenRevoked(err) {
		t.Errorf("err = %v, want TOKEN_REVOKED", err)
	}
}

func TestValidateDeletedUser(t *testing.T) {
	tokens := newFakeTokenRepo()
	repo := happyUserRepo()
	uc := newTestUseCase(t, repo, tokens)
	pair, _ := login(t, uc)

	repo.getByID = func(context.Context, string) (*AuthUser, error) {
		return nil, userv1.ErrorUserNotFound("user not found")
	}
	if _, _, _, err := uc.Validate(context.Background(), pair.AccessToken); !userv1.IsUserNotFound(err) {
		t.Errorf("err = %v, want USER_NOT_FOUND passthrough", err)
	}
}

func TestLogout(t *testing.T) {
	tokens := newFakeTokenRepo()
	uc := newTestUseCase(t, happyUserRepo(), tokens)
	pair, refreshJTI := login(t, uc)

	if err := uc.Logout(context.Background(), pair.AccessToken, pair.RefreshToken); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	claims, err := uc.jwtManager.ParseToken(pair.AccessToken, pkgauth.TokenTypeAccess)
	if err != nil {
		t.Fatalf("parse access: %v", err)
	}
	ttl, ok := tokens.revoked[claims.ID]
	if !ok {
		t.Fatal("access token not deny-listed")
	}
	if ttl < 14*time.Minute || ttl > 15*time.Minute {
		t.Errorf("denylist ttl = %v, want ~15m", ttl)
	}
	if _, ok := tokens.refresh[refreshJTI]; ok {
		t.Error("refresh token not consumed")
	}
}

func TestLogoutInvalidAccessTokenIsNoop(t *testing.T) {
	tokens := newFakeTokenRepo()
	uc := newTestUseCase(t, happyUserRepo(), tokens)

	if err := uc.Logout(context.Background(), "garbage", ""); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if len(tokens.revoked) != 0 {
		t.Errorf("revoked = %v, want none", tokens.revoked)
	}
}
