package auth

import (
	"context"

	authv1 "kratos-template/api/auth/v1"
	"kratos-template/app/gateway/biz/model/auth"
)

func (s *AuthService) Logout(ctx context.Context, req *auth.LogoutRequest) (*auth.LogoutResponse, error) {
	resp, err := s.client.Logout(ctx, &authv1.LogoutRequest{
		AccessToken: req.AccessToken,
	})
	if err != nil {
		return nil, err
	}
	return &auth.LogoutResponse{
		Success: resp.Success,
	}, nil
}
