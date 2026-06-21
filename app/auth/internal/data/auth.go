package data

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"kratos-template/app/auth/internal/biz"
	"kratos-template/pkg/model"
)

var _ biz.AuthUserRepo = (*authUserRepo)(nil)

type authUserRepo struct {
	data *Data
}

func NewAuthUserRepo(data *Data) biz.AuthUserRepo {
	return &authUserRepo{data: data}
}

func (r *authUserRepo) GetByUsername(ctx context.Context, username string) (*biz.AuthUser, error) {
	user, err := r.data.q.User.WithContext(ctx).Where(r.data.q.User.Username.Eq(username)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}
	return toAuthUser(user), nil
}

func (r *authUserRepo) GetByID(ctx context.Context, id string) (*biz.AuthUser, error) {
	user, err := r.data.q.User.WithContext(ctx).Where(r.data.q.User.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}
	return toAuthUser(user), nil
}

func toAuthUser(user *model.User) *biz.AuthUser {
	return &biz.AuthUser{
		ID:           user.ID,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
	}
}
