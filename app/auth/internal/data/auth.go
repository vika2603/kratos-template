package data

import (
	"context"
	"kratos-template/app/auth/internal/biz"
	"kratos-template/pkg/log"

	"go.uber.org/zap"

	authv1 "kratos-template/api/auth/v1"
	userv1 "kratos-template/api/user/v1"
)

var _ biz.AuthUserRepo = (*authUserRepo)(nil)

// authUserRepo implements biz.AuthUserRepo via the user service over gRPC.
type authUserRepo struct {
	data *Data
}

func NewAuthUserRepo(data *Data) biz.AuthUserRepo {
	return &authUserRepo{data: data}
}

func (r *authUserRepo) VerifyCredentials(ctx context.Context, username, password string) (*biz.AuthUser, error) {
	resp, err := r.data.user.VerifyCredentials(ctx, &userv1.VerifyCredentialsRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, translateUserError(ctx, err)
	}
	return &biz.AuthUser{ID: resp.UserId, Username: resp.Username}, nil
}

func (r *authUserRepo) GetByID(ctx context.Context, id string) (*biz.AuthUser, error) {
	resp, err := r.data.user.GetUser(ctx, &userv1.GetUserRequest{Id: id})
	if err != nil {
		return nil, translateUserError(ctx, err)
	}
	return &biz.AuthUser{ID: resp.User.Id, Username: resp.User.Username}, nil
}

func translateUserError(ctx context.Context, err error) error {
	switch {
	case userv1.IsInvalidCredentials(err), userv1.IsUserNotFound(err):
		return authv1.ErrorInvalidCredentials("invalid credentials")
	default:
		log.WithContextLogger(ctx, log.L()).Error("user service error", zap.Error(err))
		return authv1.ErrorInternal("user service error")
	}
}
