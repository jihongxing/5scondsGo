package cache

import (
	"context"
	"fmt"

	"github.com/fiveseconds/server/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisClient 全局 Redis 客户端
var RedisClient *redis.Client

// InitRedis 初始化 Redis 连接
func InitRedis(cfg *config.RedisConfig, logger *zap.Logger) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx := context.Background()
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}

	logger.Info("Redis connected", zap.String("host", cfg.Host), zap.Int("port", cfg.Port))
	return nil
}

// CloseRedis 关闭 Redis 连接
func CloseRedis() {
	if RedisClient != nil {
		RedisClient.Close()
	}
}
