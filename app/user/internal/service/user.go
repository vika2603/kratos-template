package service

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "kratos-template/api/user/v1"
	"kratos-template/app/user/internal/biz"
)

type UserService struct {
	v1.UnimplementedUserServiceServer
	userUC *biz.UserUseCase
}

func NewUserService(uc *biz.UserUseCase) v1.UserServiceServer {
	return &UserService{userUC: uc}
}

func (s *UserService) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
	user, err := s.userUC.CreateUser(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &v1.CreateUserResponse{
		User: &v1.User{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.GetUserResponse, error) {
	user, err := s.userUC.GetUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.GetUserResponse{
		User: &v1.User{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *v1.UpdateUserRequest) (*v1.UpdateUserResponse, error) {
	user, err := s.userUC.UpdateUser(ctx, req.Id, req.Username, req.Email)
	if err != nil {
		return nil, err
	}

	return &v1.UpdateUserResponse{
		User: &v1.User{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *v1.DeleteUserRequest) (*v1.DeleteUserResponse, error) {
	if err := s.userUC.DeleteUser(ctx, req.Id); err != nil {
		return nil, err
	}

	return &v1.DeleteUserResponse{
		Success: true,
	}, nil
}

func (s *UserService) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
	users, total, err := s.userUC.ListUsers(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	userList := make([]*v1.User, 0, len(users))
	for _, user := range users {
		userList = append(userList, &v1.User{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		})
	}

	return &v1.ListUsersResponse{
		Users: userList,
		Total: total,
	}, nil
}
