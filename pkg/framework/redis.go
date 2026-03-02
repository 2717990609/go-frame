// Package framework Redis 客户端
package framework

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// RedisConfig 配置
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// NewRedis 创建 Redis 客户端
func NewRedis(cfg RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("连接 Redis 失败: %w", err)
	}
	return client, nil
}
