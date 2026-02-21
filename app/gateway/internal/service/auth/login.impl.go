package auth

import (
	"context"

	authv1 "kratos-template/api/auth/v1"
	"kratos-template/app/gateway/biz/model/auth"
)

func (s *AuthService) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	resp, err := s.client.Login(ctx, &authv1.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &auth.LoginResponse{
		AccessToken: resp.AccessToken,
		TokenType:   resp.TokenType,
		ExpiresIn:   resp.ExpiresIn,
	}, nil
}
