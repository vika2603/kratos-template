package auth

import (
	"context"
	"errors"

	"kratos-template/app/gateway/biz/model/auth"
)

func (s *AuthService) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	return nil, errors.New("not implemented")
}
