// Package redis Redis 缓存插件，以插件形式启动 Redis 连接
package redis

import (
	"context"

	"go-backend-framework/pkg/framework"
	"go-backend-framework/pkg/logger"
	"go-backend-framework/pkg/plugin"

	"go.uber.org/zap"
	"github.com/go-redis/redis/v8"
)

// Plugin Redis 插件
type Plugin struct {
	plugin.BasePlugin
	rdb    *redis.Client
	config Config
}

// Config Redis 插件配置
type Config struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// NewRedisPlugin 创建 Redis 插件
func NewRedisPlugin() *Plugin {
	return &Plugin{
		BasePlugin: *plugin.NewBasePlugin("redis", "1.0.0", "Redis 缓存连接插件", "Framework Team"),
		config: Config{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}
}

// Init 解析配置
func (p *Plugin) Init(ctx context.Context, config map[string]interface{}) error {
	if v, ok := config["addr"].(string); ok {
		p.config.Addr = v
	}
	if v, ok := config["password"].(string); ok {
		p.config.Password = v
	}
	if v := getInt(config, "db"); v >= 0 {
		p.config.DB = v
	}
	return nil
}

func getInt(m map[string]interface{}, key string) int {
	v, ok := m[key]
	if !ok || v == nil {
		return -1
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return -1
	}
}

// Start 建立 Redis 连接并注册到框架
func (p *Plugin) Start(ctx context.Context) error {
	rdb, err := framework.NewRedis(framework.RedisConfig{
		Addr:     p.config.Addr,
		Password: p.config.Password,
		DB:       p.config.DB,
	})
	if err != nil {
		return err
	}
	p.rdb = rdb
	framework.RegisterRedis(rdb)
	logger.Global().Info("Redis 插件启动成功", zap.String("addr", p.config.Addr))
	return nil
}

// Stop 关闭连接并注销
func (p *Plugin) Stop(ctx context.Context) error {
	framework.UnregisterRedis()
	if p.rdb != nil {
		_ = p.rdb.Close()
		p.rdb = nil
	}
	logger.Global().Info("Redis 插件已停止")
	return nil
}
