package config

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisPlatform 平台Redis客户端
var RedisPlatform *redis.Client

// RedisEdge 边缘Redis客户端
var RedisEdge *redis.Client

// InitRedis 初始化Redis连接
func InitRedis(cfg *RedisConfig) error {
	RedisPlatform = redis.NewClient(&redis.Options{
		Addr:         cfg.Platform.Addr(),
		Password:     cfg.Platform.Password,
		DB:           cfg.Platform.DB,
		PoolSize:     cfg.Platform.PoolSize,
		MinIdleConns: cfg.Platform.MinIdleConns,
	})

	if err := RedisPlatform.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("failed to ping platform redis: %w", err)
	}
	zap.L().Info("Platform Redis connected successfully", zap.String("addr", cfg.Platform.Addr()))

	RedisEdge = redis.NewClient(&redis.Options{
		Addr:         cfg.Edge.Addr(),
		Password:     cfg.Edge.Password,
		DB:           cfg.Edge.DB,
		PoolSize:     cfg.Edge.PoolSize,
		MinIdleConns: cfg.Edge.MinIdleConns,
	})

	if err := RedisEdge.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("failed to ping edge redis: %w", err)
	}
	zap.L().Info("Edge Redis connected successfully", zap.String("addr", cfg.Edge.Addr()))

	return nil
}
