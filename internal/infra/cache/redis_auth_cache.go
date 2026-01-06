package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"rest_api_poc/internal/domain/auth"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisAuthCache struct {
	rdb *redis.Client
}

func NewRedisAuthCache(rdb *redis.Client) *RedisAuthCache {
	return &RedisAuthCache{rdb: rdb}
}

func (c *RedisAuthCache) sessionKey(sessionID string) string { return "auth:session:" + sessionID }
func (c *RedisAuthCache) userKey(userID string) string       { return "auth:user:" + userID }

func (c *RedisAuthCache) GetSession(ctx context.Context, sessionID string) (*auth.CachedSession, bool, error) {
	val, err := c.rdb.Get(ctx, c.sessionKey(sessionID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}
		return nil, false, err
	}
	var s auth.CachedSession
	if err := json.Unmarshal([]byte(val), &s); err != nil {
		// Treat corrupt cache as miss.
		_ = c.DelSession(ctx, sessionID)
		return nil, false, nil
	}
	return &s, true, nil
}

func (c *RedisAuthCache) SetSession(ctx context.Context, sessionID string, s *auth.CachedSession, ttl time.Duration) error {
	if s == nil {
		return nil
	}
	if ttl <= 0 {
		return nil
	}
	b, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal session cache: %w", err)
	}
	return c.rdb.Set(ctx, c.sessionKey(sessionID), b, ttl).Err()
}

func (c *RedisAuthCache) DelSession(ctx context.Context, sessionID string) error {
	return c.rdb.Del(ctx, c.sessionKey(sessionID)).Err()
}

func (c *RedisAuthCache) GetUser(ctx context.Context, userID string) (*auth.CachedUser, bool, error) {
	val, err := c.rdb.Get(ctx, c.userKey(userID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}
		return nil, false, err
	}
	var u auth.CachedUser
	if err := json.Unmarshal([]byte(val), &u); err != nil {
		_ = c.DelUser(ctx, userID)
		return nil, false, nil
	}
	return &u, true, nil
}

func (c *RedisAuthCache) SetUser(ctx context.Context, userID string, u *auth.CachedUser, ttl time.Duration) error {
	if u == nil {
		return nil
	}
	if ttl <= 0 {
		return nil
	}
	b, err := json.Marshal(u)
	if err != nil {
		return fmt.Errorf("marshal user cache: %w", err)
	}
	return c.rdb.Set(ctx, c.userKey(userID), b, ttl).Err()
}

func (c *RedisAuthCache) DelUser(ctx context.Context, userID string) error {
	return c.rdb.Del(ctx, c.userKey(userID)).Err()
}


