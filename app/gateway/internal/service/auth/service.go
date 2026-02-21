package auth

import (
	authv1 "kratos-template/api/auth/v1"
	"kratos-template/app/gateway/biz/model/auth"
)

type AuthService struct {
	client authv1.AuthServiceClient
}

func NewService(client authv1.AuthServiceClient) auth.AuthService {
	return &AuthService{client: client}
}
