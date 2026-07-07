package biz

import (
	"context"
	"encoding/base64"
	"errors"
	"kratos-template/pkg/middleware/authn"
	"strconv"
	"strings"
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
	// List returns users ordered by (created_at, id), keyset-paginated:
	// only rows strictly after the cursor position when afterCreatedAt is set.
	List(ctx context.Context, afterCreatedAt time.Time, afterID string, limit int) ([]*User, error)
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

func (uc *UserUseCase) ListUsers(ctx context.Context, pageSize int32, pageToken string) ([]*User, string, error) {
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var afterCreatedAt time.Time
	var afterID string
	if pageToken != "" {
		var err error
		afterCreatedAt, afterID, err = decodePageToken(pageToken)
		if err != nil {
			return nil, "", userv1.ErrorInvalidPageToken("invalid page token")
		}
	}

	// Fetch one extra row to learn whether another page exists.
	limit := int(pageSize)
	users, err := uc.repo.List(ctx, afterCreatedAt, afterID, limit+1)
	if err != nil {
		return nil, "", err
	}

	nextToken := ""
	if len(users) > limit {
		users = users[:limit]
		last := users[limit-1]
		nextToken = encodePageToken(last.CreatedAt, last.ID)
	}
	return users, nextToken, nil
}

// Page tokens encode the last row's (created_at, id); UnixMicro matches
// PostgreSQL timestamptz precision so the cursor round-trips losslessly.
func encodePageToken(createdAt time.Time, id string) string {
	raw := strconv.FormatInt(createdAt.UnixMicro(), 10) + "|" + id
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodePageToken(token string) (time.Time, string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return time.Time{}, "", err
	}
	micros, id, ok := strings.Cut(string(raw), "|")
	if !ok || id == "" {
		return time.Time{}, "", errors.New("malformed page token")
	}
	n, err := strconv.ParseInt(micros, 10, 64)
	if err != nil {
		return time.Time{}, "", err
	}
	return time.UnixMicro(n), id, nil
}
