package server

import (
	nethttp "net/http"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
	"go.uber.org/fx"

	v1 "kratos-template/api/user/v1"
	"kratos-template/app/user/internal/conf"
)

type HTTPServerParams struct {
	fx.In
	Config      *conf.Bootstrap
	Logger      log.Logger
	UserService v1.UserServiceServer
}

func NewHTTPServer(params HTTPServerParams) (*http.Server, error) {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			logging.Server(params.Logger),
		),
	}

	if params.Config.Server.Http.Addr != "" {
		opts = append(opts, http.Address(params.Config.Server.Http.Addr))
	}

	if params.Config.Server.Http.Timeout != "" {
		timeout, err := time.ParseDuration(params.Config.Server.Http.Timeout)
		if err != nil {
			return nil, err
		}
		opts = append(opts, http.Timeout(timeout))
	}

	srv := http.NewServer(opts...)
	v1.RegisterUserServiceHTTPServer(srv, params.UserService)

	srv.HandleFunc("/healthz", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	return srv, nil
}
