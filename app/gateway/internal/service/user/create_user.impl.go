package user

import (
	"context"
	"errors"

	"kratos-template/app/gateway/biz/model/user"
)

func (s *UserService) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.CreateUserResponse, error) {
	return nil, errors.New("not implemented")
}
