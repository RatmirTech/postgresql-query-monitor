package impl

import (
	"context"
	"fmt"
	"time"

	redis_internal "github.com/dreadew/go-common/pkg/clients/redis"
	redis_config "github.com/dreadew/go-common/pkg/config/redis"

	"github.com/redis/go-redis/v9"
)

type redisClient struct {
	client *redis.Client
}

func New(config *redis_config.RedisConfig) (redis_internal.RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	return &redisClient{
		client: client,
	}, nil
}

func (r *redisClient) Close() error {
	return r.client.Close()
}

func (r *redisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisClient) SetWithTimeout(ctx context.Context, key string, value interface{}, expiration int) error {
	return r.client.Set(ctx, key, value, time.Duration(expiration)*time.Second).Err()
}

func (r *redisClient) Set(ctx context.Context, key string, value interface{}) error {
	return r.SetWithTimeout(ctx, key, value, 0)
}
