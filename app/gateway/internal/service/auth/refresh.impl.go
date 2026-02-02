package auth

import (
	"context"
	"errors"

	"kratos-template/app/gateway/biz/model/auth"
)

func (s *AuthService) Refresh(ctx context.Context, req *auth.RefreshRequest) (*auth.RefreshResponse, error) {
	return nil, errors.New("not implemented")
}
