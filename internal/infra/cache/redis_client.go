package cache

import (
	"context"
	"rest_api_poc/internal/infra/config"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg *config.CacheConfig) (*redis.Client, func(ctx context.Context) error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return rdb, func(ctx context.Context) error {
		return rdb.Close()
	}
}


