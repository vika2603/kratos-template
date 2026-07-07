package metrics

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server exposes /metrics over plain HTTP. It implements transport.Server but
// deliberately not transport.Endpointer, so it is never registered in Consul.
type Server struct {
	srv *http.Server
}

// NewServer returns a disabled (no-op) server when addr is empty.
func NewServer(addr string, g prometheus.Gatherer) *Server {
	if addr == "" {
		return &Server{}
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(g, promhttp.HandlerOpts{}))
	return &Server{srv: &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}}
}

func (s *Server) Start(context.Context) error {
	if s.srv == nil {
		return nil
	}
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if s.srv == nil {
		return nil
	}
	return s.srv.Shutdown(ctx)
}
