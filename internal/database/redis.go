package database

import (
	"context"
	"fmt"
	"time"

	"github.com/cheoscafe/backend/internal/config"
	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	Ping(ctx context.Context) error
	Close() error
	GetClient() *redis.Client
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
}

type RedisDB struct {
	client *redis.Client
}

func NewRedisConnection(cfg *config.Config) (RedisClient, error) {
	var opt *redis.Options

	// Use REDIS_URL if provided (Upstash/Render style)
	if cfg.RedisURL != "" {
		parsedOpt, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			return nil, fmt.Errorf("unable to parse Redis URL: %w", err)
		}
		opt = parsedOpt
	} else {
		// Build options from individual components
		opt = &redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		}
	}

	// Create Redis client
	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("unable to ping Redis: %w", err)
	}

	return &RedisDB{client: client}, nil
}

func (r *RedisDB) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisDB) Close() error {
	return r.client.Close()
}

func (r *RedisDB) GetClient() *redis.Client {
	return r.client
}

func (r *RedisDB) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisDB) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisDB) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}
