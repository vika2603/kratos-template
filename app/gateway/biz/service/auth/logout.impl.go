package auth

import (
	"context"
	"errors"

	"kratos-template/app/gateway/biz/model/auth"
)

func (s *AuthService) Logout(ctx context.Context, req *auth.LogoutRequest) (*auth.LogoutResponse, error) {
	return nil, errors.New("not implemented")
}
