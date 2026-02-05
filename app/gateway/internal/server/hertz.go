package server

import (
	"context"
	"net/url"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/logger/accesslog"

	"kratos-template/app/gateway/internal/conf"
	"kratos-template/pkg/log"
	"kratos-template/pkg/log/adapter"
)

type HertzServer struct {
	*server.Hertz
	addr string
}

func (s *HertzServer) Start(ctx context.Context) error {
	log.Infof("hertz server starting on %s", s.addr)
	go s.Hertz.Run()
	return nil
}

func (s *HertzServer) Stop(ctx context.Context) error {
	log.Info("hertz server stopping...")
	return s.Hertz.Shutdown(ctx)
}

func (s *HertzServer) Endpoint() (*url.URL, error) {
	return url.Parse("http://" + s.addr)
}

func NewHertzServer(cfg *conf.Bootstrap) (*HertzServer, *server.Hertz, error) {
	hlog.SetLogger(adapter.NewHertzAdapter())

	addr := "0.0.0.0:8080"
	if cfg.Server.Http.Addr != "" {
		addr = cfg.Server.Http.Addr
	}

	timeout := 30 * time.Second
	if cfg.Server.Http.Timeout != "" {
		if t, err := time.ParseDuration(cfg.Server.Http.Timeout); err == nil {
			timeout = t
		}
	}

	h := server.Default(
		server.WithHostPorts(addr),
		server.WithReadTimeout(timeout),
		server.WithWriteTimeout(timeout),
		server.WithExitWaitTime(5*time.Second),
	)

	h.Use(accesslog.New())

	registerHealthCheck(h)

	return &HertzServer{Hertz: h, addr: addr}, h, nil
}

func registerHealthCheck(h *server.Hertz) {
	h.GET("/healthz", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(consts.StatusOK, map[string]string{"status": "ok"})
	})
	h.GET("/readyz", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(consts.StatusOK, map[string]string{"status": "ok"})
	})
}
