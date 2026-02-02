package data

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"kratos-template/app/auth/internal/biz"
)

var _ biz.AuthUserRepo = (*authUserRepo)(nil)

type authUserRepo struct {
	data *Data
}

func (r *authUserRepo) GetByUsername(ctx context.Context, username string) (*biz.AuthUser, error) {
	var user User
	if err := r.data.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}
	return &biz.AuthUser{
		ID:           user.ID,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
	}, nil
}

func (r *authUserRepo) GetByID(ctx context.Context, id uint) (*biz.AuthUser, error) {
	var user User
	if err := r.data.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}
	return &biz.AuthUser{
		ID:           user.ID,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
	}, nil
}
