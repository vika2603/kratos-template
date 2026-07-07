package biz

import (
	"context"
	"kratos-template/pkg/middleware/authn"
	"testing"

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

	listOffset int
	listLimit  int
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

func (f *fakeUserRepo) List(_ context.Context, offset, limit int) ([]*User, int64, error) {
	f.listOffset = offset
	f.listLimit = limit
	return nil, 0, nil
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

func TestListUsersClamping(t *testing.T) {
	tests := []struct {
		name       string
		page, size int32
		wantOffset int
		wantLimit  int
	}{
		{"defaults", 0, 0, 0, 10},
		{"negative page", -1, 20, 0, 20},
		{"size over cap", 1, 500, 0, 100},
		{"second page", 3, 25, 50, 25},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepo{}
			uc := &UserUseCase{repo: repo}
			if _, _, err := uc.ListUsers(context.Background(), tt.page, tt.size); err != nil {
				t.Fatalf("ListUsers: %v", err)
			}
			if repo.listOffset != tt.wantOffset || repo.listLimit != tt.wantLimit {
				t.Errorf("offset/limit = %d/%d, want %d/%d",
					repo.listOffset, repo.listLimit, tt.wantOffset, tt.wantLimit)
			}
		})
	}
}
