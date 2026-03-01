package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps *redis.Client with convenience methods
type RedisClient struct {
	*redis.Client
}

func NewRedisClient(redisURL string) *RedisClient {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Unable to parse Redis URL: %v", err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}

	fmt.Println("✓ Connected to Redis")
	return &RedisClient{Client: client}
}

// Inner returns the raw *redis.Client for use with middleware
func (r *RedisClient) Inner() *redis.Client {
	return r.Client
}

// Helper functions for common cache operations (method-based)

func (r *RedisClient) SetCache(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.Client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisClient) GetCache(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

func (r *RedisClient) DelCache(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

func (r *RedisClient) SetBlacklist(ctx context.Context, token string) error {
	return r.Client.Set(ctx, "blacklist:"+token, "1", 25*time.Hour).Err()
}

func (r *RedisClient) IsBlacklisted(ctx context.Context, token string) bool {
	val, err := r.Client.Get(ctx, "blacklist:"+token).Result()
	return err == nil && val == "1"
}

// Package-level functions for backward compatibility

func Set(client *redis.Client, ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return client.Set(ctx, key, value, ttl).Err()
}

func Get(client *redis.Client, ctx context.Context, key string) (string, error) {
	return client.Get(ctx, key).Result()
}

func Del(client *redis.Client, ctx context.Context, keys ...string) error {
	return client.Del(ctx, keys...).Err()
}

func SetBlacklist(client *redis.Client, ctx context.Context, token string, ttl time.Duration) error {
	return client.Set(ctx, "blacklist:"+token, "1", ttl).Err()
}

func IsBlacklisted(client *redis.Client, ctx context.Context, token string) bool {
	val, err := client.Get(ctx, "blacklist:"+token).Result()
	return err == nil && val == "1"
}
