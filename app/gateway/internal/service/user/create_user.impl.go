package user

import (
	"context"

	userv1 "kratos-template/api/user/v1"
	"kratos-template/app/gateway/biz/model/user"
)

func (s *UserService) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.CreateUserResponse, error) {
	resp, err := s.client.CreateUser(ctx, &userv1.CreateUserRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &user.CreateUserResponse{
		User: convertUser(resp.User),
	}, nil
}
