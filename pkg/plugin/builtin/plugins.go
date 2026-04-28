// Package builtin 内置插件注册
package builtin

import (
	"context"

	"go-backend-framework/config"
	"go-backend-framework/pkg/plugin"
	"go-backend-framework/plugins/mysql"
	"go-backend-framework/plugins/redis"
	"go-backend-framework/plugins/swagger"
)

// RegisterBuiltinPlugins 注册所有内置插件
func RegisterBuiltinPlugins(manager plugin.Manager) error {
	// 注册 MySQL 插件
	if err := manager.Register(mysql.NewMySQLPlugin()); err != nil {
		return err
	}
	// 注册 Redis 插件
	if err := manager.Register(redis.NewRedisPlugin()); err != nil {
		return err
	}
	// 注册 Swagger 插件
	swaggerPlugin := swagger.NewSwaggerPlugin()
	if err := manager.Register(swaggerPlugin); err != nil {
		return err
	}

	// TODO: 注册其他内置插件
	// metricsPlugin := metrics.NewMetricsPlugin()
	// if err := manager.Register(metricsPlugin); err != nil {
	//     return err
	// }
	
	// tracingPlugin := tracing.NewTracingPlugin()
	// if err := manager.Register(tracingPlugin); err != nil {
	//     return err
	// }

	return nil
}

// LoadBuiltin 加载并启动内置插件
func LoadBuiltin(ctx context.Context, manager plugin.Manager, cfg *config.Config) error {
	// 1. 注册所有内置插件
	if err := RegisterBuiltinPlugins(manager); err != nil {
		return err
	}
	// 2. 加载插件配置（启用/禁用、配置注入）
	if err := manager.LoadConfig(GetBuiltinPluginConfigs(cfg)); err != nil {
		return err
	}
	// 3. 初始化所有已启用插件
	if err := manager.Initialize(ctx); err != nil {
		return err
	}
	// 4. 启动所有已启用插件
	if err := manager.Start(ctx); err != nil {
		return err
	}
	return nil
}

// GetBuiltinPluginConfigs 从应用配置生成内置插件配置
func GetBuiltinPluginConfigs(cfg *config.Config) map[string]plugin.PluginConfig {
	configs := map[string]plugin.PluginConfig{
		"mysql": {
			Enabled: true,
			Config: map[string]interface{}{
				"host":            cfg.MySQL.Host,
				"port":            cfg.MySQL.Port,
				"user":            cfg.MySQL.User,
				"password":        cfg.MySQL.Password,
				"database":        cfg.MySQL.Database,
				"charset":         cfg.MySQL.Charset,
				"collation":       cfg.MySQL.Collation,
				"max_open_conns":  cfg.MySQL.MaxOpenConns,
				"max_idle_conns":  cfg.MySQL.MaxIdleConns,
				"enable_sql_log":  cfg.Log.EnableSQLLog,
			},
		},
		"redis": {
			Enabled: true,
			Config: map[string]interface{}{
				"addr":     cfg.Redis.Addr,
				"password": cfg.Redis.Password,
				"db":       cfg.Redis.DB,
			},
		},
		"swagger": {
			Enabled: true,
			Config: map[string]interface{}{
				"path":        "/swagger",
				"host":        "localhost:8080",
				"title":       "API Documentation",
				"description": "Auto generated API documentation",
				"version":     "1.0.0",
			},
		},
	}
	// TODO: 其他插件配置 configs["metrics"] = plugin.PluginConfig{...}
	return configs
}