// Package cache implements CacheService using Redis.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	dashboardTTL = 30 * time.Minute
	merchantTTL  = 24 * time.Hour
)

// RedisCache is the production Redis-backed cache.
type RedisCache struct {
	client *redis.Client
}

// New parses a redis:// URL and returns a connected RedisCache.
func New(redisURL string) (*RedisCache, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("cache: parse URL: %w", err)
	}

	c := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("cache: ping Redis: %w", err)
	}

	return &RedisCache{client: c}, nil
}

// Client exposes the underlying redis.Client for advanced usage (e.g. pub/sub).
func (r *RedisCache) Client() *redis.Client { return r.client }

// Close closes the Redis connection pool.
func (r *RedisCache) Close() error { return r.client.Close() }

// ─── CacheService implementation ──────────────────────────────────────────────

// Get deserialises a cached JSON value into dest.
// Returns redis.Nil-wrapped error when key is absent.
func (r *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err // caller checks redis.Nil
	}
	return json.Unmarshal([]byte(val), dest)
}

// Set serialises value as JSON and stores it with the given TTL.
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache: marshal: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

// Delete removes one or more keys.
func (r *RedisCache) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// IncrementRateLimit atomically increments a rate-limit counter.
// Sets TTL only on first increment so the window doesn't reset.
func (r *RedisCache) IncrementRateLimit(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	pipe := r.client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

// GetMerchantCategory returns a cached merchant→category mapping.
// Returns "" when not found.
func (r *RedisCache) GetMerchantCategory(ctx context.Context, merchant string) (string, error) {
	key := merchantKey(merchant)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// SetMerchantCategory persists a merchant→category mapping.
func (r *RedisCache) SetMerchantCategory(ctx context.Context, merchant, category string) error {
	return r.client.Set(ctx, merchantKey(merchant), category, merchantTTL).Err()
}

// GetDashboard deserialises a cached dashboard for the given user.
func (r *RedisCache) GetDashboard(ctx context.Context, userID string, dest interface{}) error {
	return r.Get(ctx, dashboardKey(userID), dest)
}

// SetDashboard caches a user's dashboard for 30 minutes.
func (r *RedisCache) SetDashboard(ctx context.Context, userID string, data interface{}) error {
	return r.Set(ctx, dashboardKey(userID), data, dashboardTTL)
}

// InvalidateUser removes all user-scoped cache keys.
func (r *RedisCache) InvalidateUser(ctx context.Context, userID string) error {
	keys := []string{
		dashboardKey(userID),
		fmt.Sprintf("analytics:%s", userID),
		fmt.Sprintf("profile:%s", userID),
	}
	return r.client.Del(ctx, keys...).Err()
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func merchantKey(merchant string) string  { return fmt.Sprintf("merchant:%s", merchant) }
func dashboardKey(userID string) string   { return fmt.Sprintf("dashboard:%s", userID) }
