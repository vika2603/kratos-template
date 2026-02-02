package user

import (
	"context"
	"errors"

	"kratos-template/app/gateway/biz/model/user"
)

func (s *UserService) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (*user.DeleteUserResponse, error) {
	return nil, errors.New("not implemented")
}
