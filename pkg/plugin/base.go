// Package plugin 基础插件实现
package plugin

import (
	"context"
)

// BasePlugin 基础插件实现
type BasePlugin struct {
	name        string
	version     string
	description string
	author      string
}

// NewBasePlugin 创建基础插件
func NewBasePlugin(name, version, description, author string) *BasePlugin {
	return &BasePlugin{
		name:        name,
		version:     version,
		description: description,
		author:      author,
	}
}

// Name 获取插件名称
func (p *BasePlugin) Name() string {
	return p.name
}

// Version 获取插件版本
func (p *BasePlugin) Version() string {
	return p.version
}

// Description 获取插件描述
func (p *BasePlugin) Description() string {
	return p.description
}

// Author 获取插件作者
func (p *BasePlugin) Author() string {
	return p.author
}

// Init 初始化插件（默认实现）
func (p *BasePlugin) Init(ctx context.Context, config map[string]interface{}) error {
	return nil
}

// Start 启动插件（默认实现）
func (p *BasePlugin) Start(ctx context.Context) error {
	return nil
}

// Stop 停止插件（默认实现）
func (p *BasePlugin) Stop(ctx context.Context) error {
	return nil
}

// Routes 获取路由（默认实现）
func (p *BasePlugin) Routes() []Route {
	return []Route{}
}

// Middlewares 获取中间件（默认实现）
func (p *BasePlugin) Middlewares() []Middleware {
	return []Middleware{}
}

// Hooks 获取钩子（默认实现）
func (p *BasePlugin) Hooks() []Hook {
	return []Hook{}
}