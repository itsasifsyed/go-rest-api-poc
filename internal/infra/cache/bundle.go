package cache

import (
	"context"
	"rest_api_poc/internal/domain/auth"
	"rest_api_poc/internal/infra/config"
)

// Bundle groups all cache concerns behind a single dependency.
// Add new sub-caches here over time (e.g. Product, RateLimit, etc).
type Bundle struct {
	Auth auth.AuthCache

	closeFn func(ctx context.Context) error
}

func (b *Bundle) Close(ctx context.Context) error {
	if b == nil || b.closeFn == nil {
		return nil
	}
	return b.closeFn(ctx)
}

func NewBundle(cfg *config.CacheConfig) *Bundle {
	// Default: no caching enabled.
	if cfg == nil || !cfg.Enable {
		return &Bundle{
			Auth:    nil,
			closeFn: func(context.Context) error { return nil },
		}
	}

	rdb, closeFn := NewRedisClient(cfg)
	return &Bundle{
		Auth:    NewRedisAuthCache(rdb),
		closeFn: closeFn,
	}
}
