package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/voice-diary/backend/internal/pkg/logger"
)

// RedisClient wraps redis.Client
type RedisClient struct {
	*redis.Client
}

// ConnectRedis creates a new Redis connection
func ConnectRedis(addr string) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     "", // no password by default
		DB:           0,  // default DB
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Successfully connected to Redis")
	return &RedisClient{client}, nil
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	logger.Info("Closing Redis connection")
	return r.Client.Close()
}

// HealthCheck checks if Redis is healthy
func (r *RedisClient) HealthCheck() error {
	ctx, cancel := getContext()
	defer cancel()

	if err := r.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}
	return nil
}

func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}
