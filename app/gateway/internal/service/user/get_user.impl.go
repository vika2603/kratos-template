package user

import (
	"context"

	userv1 "kratos-template/api/user/v1"
	"kratos-template/app/gateway/biz/model/user"
)

func (s *UserService) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
	resp, err := s.client.GetUser(ctx, &userv1.GetUserRequest{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}
	return &user.GetUserResponse{
		User: convertUser(resp.User),
	}, nil
}
