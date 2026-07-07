package data

import (
	"context"
	"errors"
	"kratos-template/app/auth/internal/biz"
	"time"

	"github.com/redis/go-redis/v9"
)

var _ biz.LoginGuardRepo = (*loginGuardRepo)(nil)

type loginGuardRepo struct {
	client *redis.Client
}

func NewLoginGuardRepo(data *Data) biz.LoginGuardRepo {
	return &loginGuardRepo{client: data.redis}
}

func (r *loginGuardRepo) FailureCount(ctx context.Context, username string) (int64, error) {
	n, err := r.client.Get(ctx, loginFailKey(username)).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	return n, err
}

func (r *loginGuardRepo) RecordFailure(ctx context.Context, username string, window time.Duration) error {
	key := loginFailKey(username)
	n, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	// Fixed window: the first failure starts the clock.
	if n == 1 {
		return r.client.Expire(ctx, key, window).Err()
	}
	return nil
}

func (r *loginGuardRepo) Reset(ctx context.Context, username string) error {
	return r.client.Del(ctx, loginFailKey(username)).Err()
}

func loginFailKey(username string) string {
	return "login_fail:" + username
}
