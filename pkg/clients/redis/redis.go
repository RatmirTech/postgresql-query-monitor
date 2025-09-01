package redis

import "context"

type RedisClient interface {
	Close() error
	Get(ctx context.Context, key string) (string, error)
	SetWithTimeout(ctx context.Context, key string, value interface{}, expiration int) error
	Set(ctx context.Context, key string, value interface{}) error
}
