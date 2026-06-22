package data

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"kratos-template/app/user/internal/biz"
	"kratos-template/pkg/model"
)

var _ biz.UserRepo = (*userRepo)(nil)

type userRepo struct {
	data *Data
}

func NewUserRepo(data *Data) biz.UserRepo {
	return &userRepo{data: data}
}

func (r *userRepo) Create(ctx context.Context, user *biz.User) error {
	m := toModel(user)
	if err := r.data.q.User.WithContext(ctx).Create(m); err != nil {
		return err
	}
	// Copy back DB-generated fields (id, timestamps).
	user.ID = m.ID
	user.CreatedAt = m.CreatedAt
	user.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *userRepo) GetByID(ctx context.Context, id string) (*biz.User, error) {
	user, err := r.data.q.User.WithContext(ctx).Where(r.data.q.User.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}
	return toBiz(user), nil
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*biz.User, error) {
	user, err := r.data.q.User.WithContext(ctx).Where(r.data.q.User.Username.Eq(username)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}
	return toBiz(user), nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*biz.User, error) {
	user, err := r.data.q.User.WithContext(ctx).Where(r.data.q.User.Email.Eq(email)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}
	return toBiz(user), nil
}

func (r *userRepo) Update(ctx context.Context, user *biz.User) error {
	_, err := r.data.q.User.WithContext(ctx).Where(r.data.q.User.ID.Eq(user.ID)).Updates(toModel(user))
	return err
}

func (r *userRepo) Delete(ctx context.Context, id string) error {
	_, err := r.data.q.User.WithContext(ctx).Where(r.data.q.User.ID.Eq(id)).Delete()
	return err
}

func (r *userRepo) List(ctx context.Context, offset, limit int) ([]*biz.User, int64, error) {
	users, count, err := r.data.q.User.WithContext(ctx).FindByPage(offset, limit)
	if err != nil {
		return nil, 0, err
	}
	result := make([]*biz.User, 0, len(users))
	for _, u := range users {
		result = append(result, toBiz(u))
	}
	return result, count, nil
}

func toBiz(m *model.User) *biz.User {
	return &biz.User{
		ID:           m.ID,
		Username:     m.Username,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func toModel(u *biz.User) *model.User {
	return &model.User{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}
