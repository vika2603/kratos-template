package data

import (
	"context"
	"errors"
	"kratos-template/app/auth/internal/biz"
	"time"

	"github.com/redis/go-redis/v9"
)

var _ biz.TokenRepo = (*tokenRepo)(nil)

type tokenRepo struct {
	client *redis.Client
}

func NewTokenRepo(data *Data) biz.TokenRepo {
	return &tokenRepo{client: data.redis}
}

func (r *tokenRepo) RevokeAccess(ctx context.Context, jti string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	return r.client.Set(ctx, denylistKey(jti), "1", ttl).Err()
}

func (r *tokenRepo) IsAccessRevoked(ctx context.Context, jti string) (bool, error) {
	n, err := r.client.Exists(ctx, denylistKey(jti)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *tokenRepo) SaveRefresh(ctx context.Context, jti, userID string, ttl time.Duration) error {
	pipe := r.client.Pipeline()
	setKey := userRefreshKey(userID)
	pipe.Set(ctx, refreshKey(jti), userID, ttl)
	pipe.SAdd(ctx, setKey, jti)
	pipe.Expire(ctx, setKey, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *tokenRepo) ConsumeRefresh(ctx context.Context, jti string) (string, bool, error) {
	userID, err := r.client.GetDel(ctx, refreshKey(jti)).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if err := r.client.SRem(ctx, userRefreshKey(userID), jti).Err(); err != nil {
		return "", false, err
	}
	return userID, true, nil
}

func (r *tokenRepo) RevokeAllRefresh(ctx context.Context, userID string) error {
	setKey := userRefreshKey(userID)
	jtis, err := r.client.SMembers(ctx, setKey).Result()
	if err != nil {
		return err
	}
	keys := make([]string, 0, len(jtis)+1)
	for _, jti := range jtis {
		keys = append(keys, refreshKey(jti))
	}
	keys = append(keys, setKey)
	return r.client.Del(ctx, keys...).Err()
}

func denylistKey(jti string) string {
	return "denylist:" + jti
}

func refreshKey(jti string) string {
	return "refresh:" + jti
}

func userRefreshKey(userID string) string {
	return "user_refresh:" + userID
}
