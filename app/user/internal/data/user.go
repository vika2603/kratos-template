package data

import (
	"context"
	"errors"
	"kratos-template/app/user/internal/biz"
	"kratos-template/app/user/internal/data/query"
	"kratos-template/pkg/log"
	"kratos-template/pkg/model"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"gorm.io/gorm"

	userv1 "kratos-template/api/user/v1"
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
		return translateDBError(ctx, err)
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
		return nil, translateDBError(ctx, err)
	}
	return toBiz(user), nil
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*biz.User, error) {
	user, err := r.data.q.User.WithContext(ctx).Where(r.data.q.User.Username.Eq(username)).First()
	if err != nil {
		return nil, translateDBError(ctx, err)
	}
	return toBiz(user), nil
}

func (r *userRepo) Update(ctx context.Context, user *biz.User) error {
	err := r.data.q.Transaction(func(tx *query.Query) error {
		u := tx.User
		info, err := u.WithContext(ctx).
			Where(u.ID.Eq(user.ID)).
			Updates(map[string]any{
				"username": user.Username,
				"email":    user.Email,
			})
		if err != nil {
			return translateDBError(ctx, err)
		}
		if info.RowsAffected == 0 {
			return userv1.ErrorUserNotFound("user not found")
		}
		updated, err := u.WithContext(ctx).Where(u.ID.Eq(user.ID)).First()
		if err != nil {
			return translateDBError(ctx, err)
		}
		*user = *toBiz(updated)
		return nil
	})
	return err
}

func (r *userRepo) Delete(ctx context.Context, id string) error {
	info, err := r.data.q.User.WithContext(ctx).Where(r.data.q.User.ID.Eq(id)).Delete()
	if err != nil {
		return translateDBError(ctx, err)
	}
	if info.RowsAffected == 0 {
		return userv1.ErrorUserNotFound("user not found")
	}
	return nil
}

func (r *userRepo) List(ctx context.Context, afterCreatedAt time.Time, afterID string, limit int) ([]*biz.User, error) {
	u := r.data.q.User
	do := u.WithContext(ctx)
	if !afterCreatedAt.IsZero() {
		// Keyset over (created_at, id); backed by users_created_at_id_idx.
		do = do.Where(u.CreatedAt.Gt(afterCreatedAt)).
			Or(u.CreatedAt.Eq(afterCreatedAt), u.ID.Gt(afterID))
	}
	users, err := do.Order(u.CreatedAt, u.ID).Limit(limit).Find()
	if err != nil {
		return nil, translateDBError(ctx, err)
	}
	result := make([]*biz.User, 0, len(users))
	for _, m := range users {
		result = append(result, toBiz(m))
	}
	return result, nil
}

func translateDBError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return userv1.ErrorUserNotFound("user not found")
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "users_username_key":
			return userv1.ErrorUsernameExists("username already exists")
		case "users_email_key":
			return userv1.ErrorEmailExists("email already exists")
		}
	}
	log.WithContextLogger(ctx, log.L()).Error("database error", zap.Error(err))
	return userv1.ErrorInternal("internal error")
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
