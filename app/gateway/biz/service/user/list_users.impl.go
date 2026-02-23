package user

import (
	"context"
	"errors"

	"kratos-template/app/gateway/biz/model/user"
)

func (s *UserService) ListUsers(ctx context.Context, req *user.ListUsersRequest) (*user.ListUsersResponse, error) {
	return nil, errors.New("not implemented")
}
