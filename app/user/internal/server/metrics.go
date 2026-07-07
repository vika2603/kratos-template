package server

import (
	"kratos-template/app/user/internal/conf"
	pkgmetrics "kratos-template/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

func NewMetricsServer(cfg *conf.Bootstrap, registry *prometheus.Registry) *pkgmetrics.Server {
	return pkgmetrics.NewServer(cfg.GetServer().GetMetrics().GetAddr(), registry)
}
