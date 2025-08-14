package cache

import (
	"context"
	"fmt"
	"time"

	"go-starter/config"

	"github.com/go-redis/redis/v8"
)

// NewRedisClient creates a new Redis client from configuration
func NewRedisClient(cfg *config.RedisConfig) (*redis.Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("redis configuration is required")
	}

	// Create Redis client options
	options := &redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	// Create Redis client
	client := redis.NewClient(options)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

// NewCache creates a new cache instance with Redis client
func NewCache(cfg *config.RedisConfig) (Cache, error) {
	client, err := NewRedisClient(cfg)
	if err != nil {
		return nil, err
	}

	cacheConfig := DefaultCacheConfig()
	return NewRedisCache(client, cacheConfig), nil
}
