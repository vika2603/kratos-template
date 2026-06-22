package data

import (
	"context"

	userv1 "kratos-template/api/user/v1"
	"kratos-template/app/auth/internal/biz"
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
		return nil, err
	}
	return &biz.AuthUser{ID: resp.UserId, Username: resp.Username}, nil
}

func (r *authUserRepo) GetByID(ctx context.Context, id string) (*biz.AuthUser, error) {
	resp, err := r.data.user.GetUser(ctx, &userv1.GetUserRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return &biz.AuthUser{ID: resp.User.Id, Username: resp.User.Username}, nil
}
