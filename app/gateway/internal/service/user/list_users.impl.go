package user

import (
	"context"

	userv1 "kratos-template/api/user/v1"
	"kratos-template/app/gateway/biz/model/user"
)

func (s *UserService) ListUsers(ctx context.Context, req *user.ListUsersRequest) (*user.ListUsersResponse, error) {
	resp, err := s.client.ListUsers(ctx, &userv1.ListUsersRequest{
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	users := make([]*user.User, 0, len(resp.Users))
	for _, u := range resp.Users {
		users = append(users, convertUser(u))
	}
	return &user.ListUsersResponse{
		Users: users,
		Total: int64(resp.Total),
	}, nil
}
