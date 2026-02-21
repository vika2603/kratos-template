package service

import (
	"context"

	"go.uber.org/fx"

	v1 "kratos-template/api/auth/v1"
	"kratos-template/app/auth/internal/biz"
)

type AuthService struct {
	v1.UnimplementedAuthServiceServer
	authUC *biz.AuthUseCase
}

type AuthServiceParams struct {
	fx.In
	AuthUseCase *biz.AuthUseCase
}

type AuthServiceResult struct {
	fx.Out
	AuthService v1.AuthServiceServer
}

func NewAuthService(params AuthServiceParams) AuthServiceResult {
	return AuthServiceResult{
		AuthService: &AuthService{
			authUC: params.AuthUseCase,
		},
	}
}

func (s *AuthService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	token, expiresIn, err := s.authUC.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	return &v1.LoginReply{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, req *v1.RefreshRequest) (*v1.RefreshReply, error) {
	token, expiresIn, err := s.authUC.Refresh(ctx, req.AccessToken)
	if err != nil {
		return nil, err
	}

	return &v1.RefreshReply{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
	}, nil
}

func (s *AuthService) Validate(ctx context.Context, req *v1.ValidateRequest) (*v1.ValidateReply, error) {
	valid, userID, username, err := s.authUC.Validate(ctx, req.AccessToken)
	if err != nil {
		return nil, err
	}

	return &v1.ValidateReply{
		Valid:    valid,
		UserId:   userID,
		Username: username,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, req *v1.LogoutRequest) (*v1.LogoutReply, error) {
	if err := s.authUC.Logout(ctx, req.AccessToken); err != nil {
		return nil, err
	}

	return &v1.LogoutReply{
		Success: true,
	}, nil
}
