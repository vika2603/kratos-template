package service

import (
	"context"
	"kratos-template/app/auth/internal/biz"

	v1 "kratos-template/api/auth/v1"
)

type AuthService struct {
	v1.UnimplementedAuthServiceServer
	authUC *biz.AuthUseCase
}

func NewAuthService(uc *biz.AuthUseCase) v1.AuthServiceServer {
	return &AuthService{authUC: uc}
}

func (s *AuthService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	pair, err := s.authUC.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	return &v1.LoginResponse{
		TokenPair: toProtoTokenPair(pair),
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, req *v1.RefreshRequest) (*v1.RefreshResponse, error) {
	pair, err := s.authUC.Refresh(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &v1.RefreshResponse{
		TokenPair: toProtoTokenPair(pair),
	}, nil
}

func (s *AuthService) Validate(ctx context.Context, req *v1.ValidateRequest) (*v1.ValidateResponse, error) {
	valid, userID, username, err := s.authUC.Validate(ctx, req.AccessToken)
	if err != nil {
		return nil, err
	}

	return &v1.ValidateResponse{
		Valid:    valid,
		UserId:   userID,
		Username: username,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, req *v1.LogoutRequest) (*v1.LogoutResponse, error) {
	if err := s.authUC.Logout(ctx, req.AccessToken, req.RefreshToken); err != nil {
		return nil, err
	}

	return &v1.LogoutResponse{
		Success: true,
	}, nil
}

func toProtoTokenPair(pair *biz.TokenPair) *v1.TokenPair {
	return &v1.TokenPair{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		TokenType:        pair.TokenType,
		ExpiresIn:        pair.ExpiresIn,
		RefreshExpiresIn: pair.RefreshExpiresIn,
	}
}
