package user

import (
	"context"
	"errors"

	"kratos-template/app/gateway/biz/model/user"
)

func (s *UserService) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*user.UpdateUserResponse, error) {
	return nil, errors.New("not implemented")
}
