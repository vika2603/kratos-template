package user

import (
	"context"
	"errors"

	"kratos-template/app/gateway/biz/model/user"
)

func (s *UserService) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
	return nil, errors.New("not implemented")
}
