package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"go.uber.org/fx"

	"kratos-template/app/worker/internal/conf"
)

type HTTPServerParams struct {
	fx.In
	Config *conf.Bootstrap
	Logger log.Logger
}

func NewHTTPServer(params HTTPServerParams) (*khttp.Server, error) {
	var opts = []khttp.ServerOption{}

	if params.Config.Server.Http.Addr != "" {
		opts = append(opts, khttp.Address(params.Config.Server.Http.Addr))
	}

	if params.Config.Server.Http.Timeout != "" {
		timeout, err := time.ParseDuration(params.Config.Server.Http.Timeout)
		if err != nil {
			return nil, err
		}
		opts = append(opts, khttp.Timeout(timeout))
	}

	srv := khttp.NewServer(opts...)

	srv.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "healthy",
			"service": "worker",
			"time":    time.Now().Unix(),
		})
	})

	return srv, nil
}
