package data

import (
	"context"
	"fmt"
	"kratos-template/pkg/bootstrap"

	"github.com/redis/go-redis/v9"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func NewRedisHealthChecker(client *redis.Client) bootstrap.HealthChecker {
	return bootstrap.HealthChecker{
		Name: "redis",
		Check: func(ctx context.Context) error {
			return client.Ping(ctx).Err()
		},
	}
}

// Auth can't do anything without user, so user's readiness is part of ours.
func NewUserServiceHealthChecker(conn *ggrpc.ClientConn) bootstrap.HealthChecker {
	hc := grpc_health_v1.NewHealthClient(conn)
	return bootstrap.HealthChecker{
		Name: "user-service",
		Check: func(ctx context.Context) error {
			resp, err := hc.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
			if err != nil {
				return err
			}
			if resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
				return fmt.Errorf("user service reports %s", resp.GetStatus())
			}
			return nil
		},
	}
}
