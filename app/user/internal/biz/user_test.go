package biz

import (
	"context"
	"encoding/base64"
	"kratos-template/pkg/middleware/authn"
	"testing"
	"time"

	kratosErrors "github.com/go-kratos/kratos/v2/errors"
	"golang.org/x/crypto/bcrypt"

	userv1 "kratos-template/api/user/v1"
	pkgauth "kratos-template/pkg/auth"
)

type fakeUserRepo struct {
	created *User
	updated *User
	deleted string

	createErr error
	getByName func(username string) (*User, error)

	listAfterCreatedAt time.Time
	listAfterID        string
	listLimit          int
	listResult         []*User
}

func (f *fakeUserRepo) Create(_ context.Context, user *User) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = user
	return nil
}

func (f *fakeUserRepo) GetByID(_ context.Context, id string) (*User, error) {
	return &User{ID: id, Username: "alice", Email: "alice@example.com"}, nil
}

func (f *fakeUserRepo) GetByUsername(_ context.Context, username string) (*User, error) {
	return f.getByName(username)
}

func (f *fakeUserRepo) Update(_ context.Context, user *User) error {
	f.updated = user
	return nil
}

func (f *fakeUserRepo) Delete(_ context.Context, id string) error {
	f.deleted = id
	return nil
}

func (f *fakeUserRepo) List(_ context.Context, afterCreatedAt time.Time, afterID string, limit int) ([]*User, error) {
	f.listAfterCreatedAt = afterCreatedAt
	f.listAfterID = afterID
	f.listLimit = limit
	if len(f.listResult) > limit {
		return f.listResult[:limit], nil
	}
	return f.listResult, nil
}

func ownerCtx(userID string) context.Context {
	return authn.NewContext(context.Background(), &pkgauth.Claims{UserID: userID})
}

func TestCreateUserHashesPassword(t *testing.T) {
	repo := &fakeUserRepo{}
	uc := &UserUseCase{repo: repo}

	user, err := uc.CreateUser(context.Background(), "alice", "alice@example.com", "password123")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if repo.created == nil {
		t.Fatal("repo.Create not called")
	}
	if user.PasswordHash == "password123" || user.PasswordHash == "" {
		t.Fatal("password not hashed")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123")); err != nil {
		t.Errorf("hash does not verify: %v", err)
	}
}

func TestCreateUserPasswordTooLong(t *testing.T) {
	uc := &UserUseCase{repo: &fakeUserRepo{}}
	long := make([]byte, 73)
	for i := range long {
		long[i] = 'a'
	}
	_, err := uc.CreateUser(context.Background(), "alice", "alice@example.com", string(long))
	if kratosErrors.Code(err) != 400 {
		t.Errorf("err = %v, want 400", err)
	}
}

func TestCreateUserConflictPassthrough(t *testing.T) {
	repo := &fakeUserRepo{createErr: userv1.ErrorUsernameExists("username already exists")}
	uc := &UserUseCase{repo: repo}

	_, err := uc.CreateUser(context.Background(), "alice", "alice@example.com", "password123")
	if !userv1.IsUsernameExists(err) {
		t.Errorf("err = %v, want USERNAME_EXISTS passthrough", err)
	}
}

// Missing user and wrong password must be indistinguishable to callers.
func TestVerifyCredentialsUniformError(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}

	tests := []struct {
		name      string
		getByName func(string) (*User, error)
	}{
		{"user not found", func(string) (*User, error) {
			return nil, userv1.ErrorUserNotFound("user not found")
		}},
		{"wrong password", func(string) (*User, error) {
			return &User{ID: "u1", PasswordHash: string(hash)}, nil
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &UserUseCase{repo: &fakeUserRepo{getByName: tt.getByName}}
			_, err := uc.VerifyCredentials(context.Background(), "alice", "not-the-password")
			if !userv1.IsInvalidCredentials(err) {
				t.Errorf("err = %v, want INVALID_CREDENTIALS", err)
			}
		})
	}
}

func TestVerifyCredentialsSuccess(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	uc := &UserUseCase{repo: &fakeUserRepo{getByName: func(string) (*User, error) {
		return &User{ID: "u1", PasswordHash: string(hash)}, nil
	}}}

	user, err := uc.VerifyCredentials(context.Background(), "alice", "password123")
	if err != nil {
		t.Fatalf("VerifyCredentials: %v", err)
	}
	if user.ID != "u1" {
		t.Errorf("ID = %q, want u1", user.ID)
	}
}

func TestUpdateUserOwnership(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{"owner", ownerCtx("u1"), false},
		{"other user", ownerCtx("u2"), true},
		{"no claims", context.Background(), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepo{}
			uc := &UserUseCase{repo: repo}
			_, err := uc.UpdateUser(tt.ctx, "u1", "newname", "")
			if tt.wantErr {
				if !userv1.IsPermissionDenied(err) {
					t.Errorf("err = %v, want PERMISSION_DENIED", err)
				}
				if repo.updated != nil {
					t.Error("repo.Update called despite denial")
				}
			} else {
				if err != nil {
					t.Fatalf("UpdateUser: %v", err)
				}
				if repo.updated == nil || repo.updated.Username != "newname" {
					t.Errorf("updated = %+v, want username newname", repo.updated)
				}
			}
		})
	}
}

func TestDeleteUserOwnership(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{"owner", ownerCtx("u1"), false},
		{"other user", ownerCtx("u2"), true},
		{"no claims", context.Background(), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepo{}
			uc := &UserUseCase{repo: repo}
			err := uc.DeleteUser(tt.ctx, "u1")
			if tt.wantErr {
				if !userv1.IsPermissionDenied(err) {
					t.Errorf("err = %v, want PERMISSION_DENIED", err)
				}
				if repo.deleted != "" {
					t.Error("repo.Delete called despite denial")
				}
			} else {
				if err != nil {
					t.Fatalf("DeleteUser: %v", err)
				}
				if repo.deleted != "u1" {
					t.Errorf("deleted = %q, want u1", repo.deleted)
				}
			}
		})
	}
}

func TestListUsersPageSizeClamping(t *testing.T) {
	// The repo always sees pageSize+1: the extra row detects a next page.
	tests := []struct {
		name      string
		size      int32
		wantLimit int
	}{
		{"default", 0, 11},
		{"negative", -3, 11},
		{"explicit", 20, 21},
		{"over cap", 500, 101},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepo{}
			uc := &UserUseCase{repo: repo}
			if _, _, err := uc.ListUsers(context.Background(), tt.size, ""); err != nil {
				t.Fatalf("ListUsers: %v", err)
			}
			if repo.listLimit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", repo.listLimit, tt.wantLimit)
			}
		})
	}
}

func TestListUsersPagination(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	all := []*User{
		{ID: "a", CreatedAt: base},
		{ID: "b", CreatedAt: base.Add(time.Second)},
		{ID: "c", CreatedAt: base.Add(2 * time.Second)},
	}
	repo := &fakeUserRepo{listResult: all}
	uc := &UserUseCase{repo: repo}

	users, token, err := uc.ListUsers(context.Background(), 2, "")
	if err != nil {
		t.Fatalf("first page: %v", err)
	}
	if len(users) != 2 || users[0].ID != "a" || users[1].ID != "b" {
		t.Fatalf("first page = %+v, want [a b]", users)
	}
	if token == "" {
		t.Fatal("expected non-empty next_page_token")
	}

	repo.listResult = all[2:]
	users, token, err = uc.ListUsers(context.Background(), 2, token)
	if err != nil {
		t.Fatalf("second page: %v", err)
	}
	if !repo.listAfterCreatedAt.Equal(all[1].CreatedAt) || repo.listAfterID != "b" {
		t.Errorf("cursor = (%v, %q), want position of user b", repo.listAfterCreatedAt, repo.listAfterID)
	}
	if len(users) != 1 || users[0].ID != "c" {
		t.Fatalf("second page = %+v, want [c]", users)
	}
	if token != "" {
		t.Errorf("last page token = %q, want empty", token)
	}
}

func TestListUsersInvalidPageToken(t *testing.T) {
	tokens := map[string]string{
		"not base64":   "%%%",
		"no separator": base64.RawURLEncoding.EncodeToString([]byte("12345")),
		"empty id":     base64.RawURLEncoding.EncodeToString([]byte("12345|")),
		"bad micros":   base64.RawURLEncoding.EncodeToString([]byte("abc|id")),
	}
	for name, token := range tokens {
		t.Run(name, func(t *testing.T) {
			uc := &UserUseCase{repo: &fakeUserRepo{}}
			_, _, err := uc.ListUsers(context.Background(), 10, token)
			if !userv1.IsInvalidPageToken(err) {
				t.Errorf("err = %v, want INVALID_PAGE_TOKEN", err)
			}
		})
	}
}

func TestPageTokenRoundtrip(t *testing.T) {
	at := time.Date(2026, 7, 7, 12, 34, 56, 789123000, time.UTC) // whole µs: matches pg precision
	token := encodePageToken(at, "e7b8a1c2-0000-0000-0000-000000000000")

	gotAt, gotID, err := decodePageToken(token)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !gotAt.Equal(at) || gotID != "e7b8a1c2-0000-0000-0000-000000000000" {
		t.Errorf("roundtrip = (%v, %q), want (%v, e7b8...)", gotAt, gotID, at)
	}
}
