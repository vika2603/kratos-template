package user

import (
	"context"

	userv1 "kratos-template/api/user/v1"
	"kratos-template/app/gateway/biz/model/user"
)

func (s *UserService) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*user.UpdateUserResponse, error) {
	resp, err := s.client.UpdateUser(ctx, &userv1.UpdateUserRequest{
		Id:       req.Id,
		Username: req.Username,
		Email:    req.Email,
	})
	if err != nil {
		return nil, err
	}
	return &user.UpdateUserResponse{
		User: convertUser(resp.User),
	}, nil
}
