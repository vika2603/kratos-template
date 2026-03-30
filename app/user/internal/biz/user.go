package biz

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	kratosErrors "github.com/go-kratos/kratos/v2/errors"

	"kratos-template/pkg/model"
)

type UserRepo interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*model.User, int64, error)
}

var (
	ErrUserNotFound   = kratosErrors.NotFound("USER_NOT_FOUND", "user not found")
	ErrUsernameExists = kratosErrors.Conflict("USERNAME_EXISTS", "username already exists")
	ErrEmailExists    = kratosErrors.Conflict("EMAIL_EXISTS", "email already exists")
)

type UserUseCase struct {
	repo UserRepo
}

func (uc *UserUseCase) CreateUser(ctx context.Context, username, email, password string) (*model.User, error) {
	existingUser, err := uc.repo.GetByUsername(ctx, username)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUsernameExists
	}

	existingUser, err = uc.repo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *UserUseCase) GetUser(ctx context.Context, id string) (*model.User, error) {
	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (uc *UserUseCase) UpdateUser(ctx context.Context, id, username, email string) (*model.User, error) {
	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if username != "" && username != user.Username {
		existingUser, err := uc.repo.GetByUsername(ctx, username)
		if err != nil && !errors.Is(err, ErrUserNotFound) {
			return nil, err
		}
		if existingUser != nil && existingUser.ID != id {
			return nil, ErrUsernameExists
		}
		user.Username = username
	}

	if email != "" && email != user.Email {
		existingUser, err := uc.repo.GetByEmail(ctx, email)
		if err != nil && !errors.Is(err, ErrUserNotFound) {
			return nil, err
		}
		if existingUser != nil && existingUser.ID != id {
			return nil, ErrEmailExists
		}
		user.Email = email
	}

	if err := uc.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *UserUseCase) DeleteUser(ctx context.Context, id string) error {
	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	return uc.repo.Delete(ctx, user.ID)
}

func (uc *UserUseCase) ListUsers(ctx context.Context, page, pageSize int32) ([]*model.User, int32, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := int((page - 1) * pageSize)
	limit := int(pageSize)

	users, total, err := uc.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	return users, int32(total), nil
}
