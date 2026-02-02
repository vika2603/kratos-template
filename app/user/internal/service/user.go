package service

import (
	"context"

	"go.uber.org/fx"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "kratos-template/api/user/v1"
	"kratos-template/app/user/internal/biz"
)

type UserService struct {
	v1.UnimplementedUserServiceServer
	userUC *biz.UserUseCase
}

type UserServiceParams struct {
	fx.In
	UserUseCase *biz.UserUseCase
}

type UserServiceResult struct {
	fx.Out
	UserService v1.UserServiceServer
}

func NewUserService(params UserServiceParams) UserServiceResult {
	return UserServiceResult{
		UserService: &UserService{
			userUC: params.UserUseCase,
		},
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserReply, error) {
	user, err := s.userUC.CreateUser(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		return nil, mapError(err)
	}

	return &v1.CreateUserReply{
		User: &v1.User{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.GetUserReply, error) {
	user, err := s.userUC.GetUser(ctx, req.Id)
	if err != nil {
		return nil, mapError(err)
	}

	return &v1.GetUserReply{
		User: &v1.User{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *v1.UpdateUserRequest) (*v1.UpdateUserReply, error) {
	user, err := s.userUC.UpdateUser(ctx, req.Id, req.Username, req.Email)
	if err != nil {
		return nil, mapError(err)
	}

	return &v1.UpdateUserReply{
		User: &v1.User{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *v1.DeleteUserRequest) (*v1.DeleteUserReply, error) {
	if err := s.userUC.DeleteUser(ctx, req.Id); err != nil {
		return nil, mapError(err)
	}

	return &v1.DeleteUserReply{
		Success: true,
	}, nil
}

func (s *UserService) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersReply, error) {
	users, total, err := s.userUC.ListUsers(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, mapError(err)
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

	return &v1.ListUsersReply{
		Users: userList,
		Total: total,
	}, nil
}

var Module = fx.Module("service",
	fx.Provide(NewUserService),
)
