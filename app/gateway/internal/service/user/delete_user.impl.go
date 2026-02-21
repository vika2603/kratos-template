package user

import (
	"context"

	userv1 "kratos-template/api/user/v1"
	"kratos-template/app/gateway/biz/model/user"
)

func (s *UserService) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (*user.DeleteUserResponse, error) {
	resp, err := s.client.DeleteUser(ctx, &userv1.DeleteUserRequest{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}
	return &user.DeleteUserResponse{
		Success: resp.Success,
	}, nil
}
