package biz

import "go.uber.org/fx"

func NewUserUseCase(repo UserRepo) *UserUseCase {
	return &UserUseCase{repo: repo}
}

var Module = fx.Module("user.biz",
	fx.Provide(NewUserUseCase),
)
