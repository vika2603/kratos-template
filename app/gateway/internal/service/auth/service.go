package auth

import (
	"kratos-template/app/gateway/biz/model/auth"
)

type AuthService struct{}

func NewService() auth.AuthService {
	return &AuthService{}
}
