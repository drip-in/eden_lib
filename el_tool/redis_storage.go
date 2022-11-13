package el_tool

import (
	"context"
	"github.com/go-redis/redis"
	"time"
)

type Storage struct {
	namespace   string
	redisClient *redis.Client
}

func NewStorage(namespace string, redisClient *redis.Client) IStorage {
	if redisClient == nil {
		panic("invalid cache config for poi settle cache")
	}

	return &Storage{namespace, redisClient}
}

func (s *Storage) Set(ctx context.Context, key string, val interface{}, expiration time.Duration) error {
	return s.redisClient.WithContext(ctx).Set(key, val, expiration).Err()
}

func (s *Storage) SetNX(ctx context.Context, key string, val interface{}, expiration time.Duration) error {
	return s.redisClient.WithContext(ctx).SetNX(key, val, expiration).Err()
}

func (s *Storage) Get(ctx context.Context, key string) (interface{}, error) {
	return s.redisClient.WithContext(ctx).Get(key).Result()
}

func (s *Storage) Del(ctx context.Context, keys ...string) error {
	return s.redisClient.WithContext(ctx).Del(keys...).Err()
}
