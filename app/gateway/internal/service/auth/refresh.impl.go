package auth

import (
	"context"

	authv1 "kratos-template/api/auth/v1"
	"kratos-template/app/gateway/biz/model/auth"
)

func (s *AuthService) Refresh(ctx context.Context, req *auth.RefreshRequest) (*auth.RefreshResponse, error) {
	resp, err := s.client.Refresh(ctx, &authv1.RefreshRequest{
		AccessToken: req.AccessToken,
	})
	if err != nil {
		return nil, err
	}
	return &auth.RefreshResponse{
		AccessToken: resp.AccessToken,
		TokenType:   resp.TokenType,
		ExpiresIn:   resp.ExpiresIn,
	}, nil
}
