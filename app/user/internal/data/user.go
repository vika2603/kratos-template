package data

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"kratos-template/app/user/internal/biz"
	"kratos-template/pkg/model"
)

var _ biz.UserRepo = (*UserRepo)(nil)

type UserRepo struct {
	data *Data
}

func NewUserRepo(data *Data) biz.UserRepo {
	return &UserRepo{data: data}
}

func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	return r.data.q.User.WithContext(ctx).Create(user)
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	user, err := r.data.q.User.WithContext(ctx).Where(r.data.q.User.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	user, err := r.data.q.User.WithContext(ctx).Where(r.data.q.User.Username.Eq(username)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := r.data.q.User.WithContext(ctx).Where(r.data.q.User.Email.Eq(email)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) Update(ctx context.Context, user *model.User) error {
	_, err := r.data.q.User.WithContext(ctx).Where(r.data.q.User.ID.Eq(user.ID)).Updates(user)
	return err
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	_, err := r.data.q.User.WithContext(ctx).Where(r.data.q.User.ID.Eq(id)).Delete()
	return err
}

func (r *UserRepo) List(ctx context.Context, offset, limit int) ([]*model.User, int64, error) {
	u := r.data.q.User
	users, count, err := u.WithContext(ctx).FindByPage(offset, limit)
	return users, count, err
}
