package user

import (
	userv1 "kratos-template/api/user/v1"
	"kratos-template/app/gateway/biz/model/user"
)

type UserService struct {
	client userv1.UserServiceClient
}

func NewService(client userv1.UserServiceClient) user.UserService {
	return &UserService{client: client}
}

func convertUser(u *userv1.User) *user.User {
	if u == nil {
		return nil
	}
	return &user.User{
		Id:        u.Id,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
