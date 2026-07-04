package data

import (
	"cmp"
	"context"
	"fmt"
	"kratos-template/app/auth/internal/conf"
	"kratos-template/pkg/log"
	"kratos-template/pkg/middleware/authn"
	"os"
	"time"

	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/registry"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/redis/go-redis/v9"
	ggrpc "google.golang.org/grpc"

	userv1 "kratos-template/api/user/v1"

	pkgauth "kratos-template/pkg/auth"
)

// Data's only resource is a gRPC client to the user service — auth owns no DB.
type Data struct {
	user  userv1.UserServiceClient
	redis *redis.Client
}

// NewUserClientConn dials the user service (discovery used when a registry exists).
func NewUserClientConn(cfg *conf.Bootstrap, disc registry.Discovery) (*ggrpc.ClientConn, error) {
	endpoint := cmp.Or(os.Getenv("USER_SERVICE_ENDPOINT"), cfg.GetData().GetUserService().GetEndpoint())
	manager, err := pkgauth.NewJWTManager(
		cmp.Or(os.Getenv("JWT_SECRET"), cfg.GetAuth().GetJwtSecret()),
		0,
		0,
	)
	if err != nil {
		return nil, err
	}

	opts := []kgrpc.ClientOption{
		kgrpc.WithEndpoint(endpoint),
		kgrpc.WithTimeout(5 * time.Second),
		kgrpc.WithMiddleware(
			tracing.Client(),
			authn.ClientServiceToken(manager, "auth"),
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

func NewRedisClient(cfg *conf.Bootstrap) (*redis.Client, error) {
	redisCfg := cfg.GetData().GetRedis()
	client := redis.NewClient(&redis.Options{
		Addr:     cmp.Or(os.Getenv("REDIS_ADDR"), redisCfg.GetAddr()),
		Password: redisCfg.GetPassword(),
		DB:       int(redisCfg.GetDb()),
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed pinging redis: %w", err)
	}
	return client, nil
}

func NewData(conn *ggrpc.ClientConn, redisClient *redis.Client) (*Data, func(), error) {
	d := &Data{
		user:  userv1.NewUserServiceClient(conn),
		redis: redisClient,
	}

	cleanup := func() {
		log.Info("closing data resources")
		if err := conn.Close(); err != nil {
			log.Errorf("failed to close user service conn: %v", err)
		}
		if err := redisClient.Close(); err != nil {
			log.Errorf("failed to close redis client: %v", err)
		}
	}

	return d, cleanup, nil
}
