package user

import "kratos-template/app/gateway/biz/model/user"

type UserService struct{}

func NewService() user.UserService {
	return &UserService{}
}
