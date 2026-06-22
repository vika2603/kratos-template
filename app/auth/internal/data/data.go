package data

import (
	"context"
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/registry"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	ggrpc "google.golang.org/grpc"

	userv1 "kratos-template/api/user/v1"
	"kratos-template/app/auth/internal/conf"
	"kratos-template/pkg/log"
)

// Data's only resource is a gRPC client to the user service — auth owns no DB.
type Data struct {
	user userv1.UserServiceClient
}

// NewUserClientConn dials the user service (discovery used when a registry exists).
func NewUserClientConn(cfg *conf.Bootstrap, disc registry.Discovery) (*ggrpc.ClientConn, error) {
	endpoint := os.Getenv("USER_SERVICE_ENDPOINT")
	if endpoint == "" {
		endpoint = cfg.GetData().GetUserService().GetEndpoint()
	}

	opts := []kgrpc.ClientOption{
		kgrpc.WithEndpoint(endpoint),
		kgrpc.WithMiddleware(
			recovery.Recovery(),
			tracing.Client(),
		),
	}
	if disc != nil {
		opts = append(opts, kgrpc.WithDiscovery(disc))
	}

	conn, err := kgrpc.DialInsecure(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed dialing user service at %q: %w", endpoint, err)
	}
	return conn, nil
}

func NewData(conn *ggrpc.ClientConn) (*Data, func(), error) {
	d := &Data{
		user: userv1.NewUserServiceClient(conn),
	}

	cleanup := func() {
		log.Info("closing data resources")
		if err := conn.Close(); err != nil {
			log.Errorf("failed to close user service conn: %v", err)
		}
	}

	return d, cleanup, nil
}
