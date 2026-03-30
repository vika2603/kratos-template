package router

import (
	auth "kratos-template/app/gateway/biz/router/auth"
	echo "kratos-template/app/gateway/biz/router/echo"
	user "kratos-template/app/gateway/biz/router/user"

	"go.uber.org/fx"
)

func Options() fx.Option {
	r := make(map[string]fx.Option)

	// INSERT_POINT: DO NOT DELETE THIS LINE!
	user.Register(r)

	echo.Register(r)

	auth.Register(r)

	var opts []fx.Option
	for _, v := range r {
		opts = append(opts, v)
	}
	return fx.Options(opts...)
}
