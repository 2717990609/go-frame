// Package swagger Swagger文档插件
package swagger

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go-backend-framework/pkg/plugin"

	swaggerFiles "github.com/swaggo/files"
)

// SwaggerPlugin Swagger插件
type SwaggerPlugin struct {
	plugin.BasePlugin
	config Config
}

// Config Swagger插件配置
type Config struct {
	Enabled     bool   `yaml:"enabled"`
	Path        string `yaml:"path"`
	Host        string `yaml:"host"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
}

// NewSwaggerPlugin 创建Swagger插件
func NewSwaggerPlugin() *SwaggerPlugin {
	return &SwaggerPlugin{
		BasePlugin: *plugin.NewBasePlugin("swagger", "1.0.0", "Swagger API文档插件", "Framework Team"),
		config: Config{
			Enabled:     true,
			Path:        "/swagger",
			Host:        "localhost:8080",
			Title:       "API文档",
			Description: "Auto generated API documentation",
			Version:     "1.0.0",
		},
	}
}

// Init 初始化插件
func (p *SwaggerPlugin) Init(ctx context.Context, config map[string]interface{}) error {
	// 解析配置
	if enabled, ok := config["enabled"].(bool); ok {
		p.config.Enabled = enabled
	}
	if path, ok := config["path"].(string); ok {
		p.config.Path = path
	}
	if host, ok := config["host"].(string); ok {
		p.config.Host = host
	}
	if title, ok := config["title"].(string); ok {
		p.config.Title = title
	}
	if description, ok := config["description"].(string); ok {
		p.config.Description = description
	}
	if version, ok := config["version"].(string); ok {
		p.config.Version = version
	}

	return p.ValidateConfig(config)
}

// Name 获取插件名称
func (p *SwaggerPlugin) Name() string {
	return "swagger"
}

// Version 获取插件版本
func (p *SwaggerPlugin) Version() string {
	return "1.0.0"
}

// Description 获取插件描述
func (p *SwaggerPlugin) Description() string {
	return "Swagger API文档生成插件"
}

// Author 获取插件作者
func (p *SwaggerPlugin) Author() string {
	return "Framework Team"
}

// Start 启动插件
func (p *SwaggerPlugin) Start(ctx context.Context) error {
	if !p.config.Enabled {
		return nil
	}

	// 这里可以添加Swagger初始化逻辑
	return nil
}

// Stop 停止插件
func (p *SwaggerPlugin) Stop(ctx context.Context) error {
	return nil
}

// Routes 获取路由
func (p *SwaggerPlugin) Routes() []plugin.Route {
	if !p.config.Enabled {
		return []plugin.Route{}
	}

	return []plugin.Route{
		{
			Method:      "GET",
			Path:        fmt.Sprintf("%s/*any", p.config.Path),
			Handler:     p.swaggerHandler(),
			Description: "Swagger UI文档",
			Tags:        []string{"docs"},
		},
	}
}

// ValidateConfig 校验配置
func (p *SwaggerPlugin) ValidateConfig(config map[string]interface{}) error {
	if p.config.Path == "" {
		return fmt.Errorf("swagger.path 不能为空")
	}
	if !strings.HasPrefix(p.config.Path, "/") {
		return fmt.Errorf("swagger.path 必须以 / 开头")
	}
	return nil
}

// Middlewares 获取中间件
func (p *SwaggerPlugin) Middlewares() []plugin.Middleware {
	return []plugin.Middleware{}
}

// Hooks 获取钩子
func (p *SwaggerPlugin) Hooks() []plugin.Hook {
	return []plugin.Hook{}
}

// swaggerHandler Swagger处理器
func (p *SwaggerPlugin) swaggerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		swaggerFiles.Handler.ServeHTTP(w, r)
	}
}

// GetConfig 获取配置
func (p *SwaggerPlugin) GetConfig() Config {
	return p.config
}

// SetConfig 设置配置
func (p *SwaggerPlugin) SetConfig(config Config) {
	p.config = config
}
