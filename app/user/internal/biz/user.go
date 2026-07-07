package biz

import (
	"context"
	"kratos-template/pkg/middleware/authn"
	"time"

	kratosErrors "github.com/go-kratos/kratos/v2/errors"
	"golang.org/x/crypto/bcrypt"

	userv1 "kratos-template/api/user/v1"
)

// User is the biz-layer domain type; the data layer maps it to/from pkg/model.User.
type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserRepo interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*User, int64, error)
}

type UserUseCase struct {
	repo UserRepo
}

func (uc *UserUseCase) CreateUser(ctx context.Context, username, email, password string) (*User, error) {
	if len(password) > 72 {
		return nil, kratosErrors.BadRequest("VALIDATION_FAILED", "password must be at most 72 bytes")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// VerifyCredentials checks a password. Same error for missing user or bad
// password, so callers can't tell which.
func (uc *UserUseCase) VerifyCredentials(ctx context.Context, username, password string) (*User, error) {
	user, err := uc.repo.GetByUsername(ctx, username)
	if err != nil {
		if userv1.IsUserNotFound(err) {
			return nil, userv1.ErrorInvalidCredentials("invalid credentials")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, userv1.ErrorInvalidCredentials("invalid credentials")
	}

	return user, nil
}

func (uc *UserUseCase) GetUser(ctx context.Context, id string) (*User, error) {
	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (uc *UserUseCase) UpdateUser(ctx context.Context, id, username, email string) (*User, error) {
	if err := requireOwner(ctx, id); err != nil {
		return nil, err
	}

	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if username != "" && username != user.Username {
		user.Username = username
	}

	if email != "" && email != user.Email {
		user.Email = email
	}

	if err := uc.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *UserUseCase) DeleteUser(ctx context.Context, id string) error {
	if err := requireOwner(ctx, id); err != nil {
		return err
	}
	return uc.repo.Delete(ctx, id)
}

// requireOwner fails closed when claims are absent (e.g. middleware removed).
func requireOwner(ctx context.Context, id string) error {
	claims, ok := authn.FromContext(ctx)
	if !ok || claims.UserID != id {
		return userv1.ErrorPermissionDenied("cannot modify another user")
	}
	return nil
}

func (uc *UserUseCase) ListUsers(ctx context.Context, page, pageSize int32) ([]*User, int32, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := int(int64(page-1) * int64(pageSize))
	limit := int(pageSize)

	users, total, err := uc.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	if total > maxListTotal {
		total = maxListTotal
	}

	return users, int32(total), nil
}

const maxListTotal = int64(1<<31 - 1)
