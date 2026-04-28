// Package plugin 插件管理器实现
package plugin

import (
	"context"
	"fmt"
	"net/http"

	"go-backend-framework/pkg/logger"
	"go.uber.org/zap"
)

// manager 插件管理器实现
type manager struct {
	registry Registry
	plugins  []Plugin
	config   map[string]PluginConfig
}

// NewManager 创建插件管理器
func NewManager() Manager {
	return &manager{
		registry: NewRegistry(),
		plugins:  make([]Plugin, 0),
		config:   make(map[string]PluginConfig),
	}
}

// Register 注册插件
func (m *manager) Register(plugin Plugin) error {
	return m.registry.Register(plugin)
}

// Unregister 注销插件
func (m *manager) Unregister(name string) error {
	return m.registry.Unregister(name)
}

// Get 获取插件
func (m *manager) Get(name string) (Plugin, bool) {
	return m.registry.Get(name)
}

// List 获取所有插件
func (m *manager) List() []Plugin {
	return m.registry.List()
}

// Enable 启用插件
func (m *manager) Enable(name string) error {
	return m.registry.Enable(name)
}

// Disable 禁用插件
func (m *manager) Disable(name string) error {
	return m.registry.Disable(name)
}

// Status 获取插件状态
func (m *manager) Status(name string) PluginStatus {
	return m.registry.Status(name)
}

// EnabledPlugins 获取已启用的插件
func (m *manager) EnabledPlugins() []Plugin {
	return m.registry.EnabledPlugins()
}

// LoadConfig 加载插件配置
func (m *manager) LoadConfig(config map[string]PluginConfig) error {
	m.config = config
	
	// 根据配置启用/禁用插件
	for pluginName, pluginConfig := range config {
		if pluginConfig.Enabled {
			if _, exists := m.registry.Get(pluginName); exists {
				if err := m.registry.Enable(pluginName); err != nil {
					logger.Global().Error("启用插件失败",
						zap.String("plugin", pluginName),
						zap.String("error", err.Error()),
					)
				}
			}
		} else {
			if err := m.registry.Disable(pluginName); err != nil {
				logger.Global().Error("禁用插件失败",
					zap.String("plugin", pluginName),
					zap.String("error", err.Error()),
				)
			}
		}
	}
	
	logger.Global().Info("插件配置加载完成",
		zap.Int("total_plugins", len(config)),
		zap.Int("enabled", m.countEnabled()),
	)
	
	return nil
}

// Initialize 初始化所有插件
func (m *manager) Initialize(ctx context.Context) error {
	plugins := m.registry.EnabledPlugins()
	
	for _, plug := range plugins {
		pluginName := plug.Name()
		
		// 获取插件配置
		config := map[string]interface{}{}
		if pluginConfig, exists := m.config[pluginName]; exists {
			config = pluginConfig.Config
		}
		
		// 初始化插件
		if err := plug.Init(ctx, config); err != nil {
			logger.Global().Error("插件初始化失败",
				zap.String("plugin", pluginName),
				zap.String("error", err.Error()),
			)
			
			// 更新状态为错误
			if reg, ok := m.registry.(*registry); ok {
				reg.setStatus(pluginName, PluginStatusError)
			}
			
			continue
		}
		
		// 更新状态为已初始化
		if reg, ok := m.registry.(*registry); ok {
			reg.setStatus(pluginName, PluginStatusEnabled)
		}
		
		logger.Global().Info("插件初始化成功",
			zap.String("plugin", pluginName),
		)
	}
	
	return nil
}

// Start 启动所有插件
func (m *manager) Start(ctx context.Context) error {
	plugins := m.registry.EnabledPlugins()
	
	for _, plug := range plugins {
		pluginName := plug.Name()
		
		// 更新状态为启动中
		if reg, ok := m.registry.(*registry); ok {
			reg.setStatus(pluginName, PluginStatusStarting)
		}
		
		// 启动插件
		if err := plug.Start(ctx); err != nil {
			logger.Global().Error("插件启动失败",
				zap.String("plugin", pluginName),
				zap.String("error", err.Error()),
			)
			
			// 更新状态为错误
			if reg, ok := m.registry.(*registry); ok {
				reg.setStatus(pluginName, PluginStatusError)
			}
			
			continue
		}
		
		// 更新状态为运行中
		if reg, ok := m.registry.(*registry); ok {
			reg.setStatus(pluginName, PluginStatusRunning)
		}
		
		// 发布启动事件
		if reg, ok := m.registry.(*registry); ok {
			reg.PublishEvent(EventPluginStarted, pluginName, "插件启动成功", nil)
		}
		
		logger.Global().Info("插件启动成功",
			zap.String("plugin", pluginName),
		)
	}
	
	return nil
}

// Stop 停止所有插件
func (m *manager) Stop(ctx context.Context) error {
	plugins := m.registry.List()
	
	// 按相反顺序停止插件
	for i := len(plugins) - 1; i >= 0; i-- {
		plug := plugins[i]
		pluginName := plug.Name()
		
		// 只停止运行中的插件
		if m.registry.Status(pluginName) != PluginStatusRunning {
			continue
		}
		
		// 更新状态为停止中
		if reg, ok := m.registry.(*registry); ok {
			reg.setStatus(pluginName, PluginStatusStopping)
		}
		
		// 停止插件
		if err := plug.Stop(ctx); err != nil {
			logger.Global().Error("插件停止失败",
				zap.String("plugin", pluginName),
				zap.String("error", err.Error()),
			)
			continue
		}
		
		// 更新状态为已启用
		if reg, ok := m.registry.(*registry); ok {
			reg.setStatus(pluginName, PluginStatusEnabled)
		}
		
		// 发布停止事件
		if reg, ok := m.registry.(*registry); ok {
			reg.PublishEvent(EventPluginStopped, pluginName, "插件停止成功", nil)
		}
		
		logger.Global().Info("插件停止成功",
			zap.String("plugin", pluginName),
		)
	}
	
	return nil
}

// ApplyRoutes 应用插件路由
func (m *manager) ApplyRoutes(mux http.Handler) http.Handler {
	plugins := m.registry.EnabledPlugins()
	
	for _, plug := range plugins {
		routes := plug.Routes()
		for _, route := range routes {
			// 这里需要集成到具体的路由框架中
			// 例如：gin、echo 等
			logger.Global().Debug("插件路由",
				zap.String("plugin", plug.Name()),
				zap.String("method", route.Method),
				zap.String("path", route.Path),
			)
		}
	}
	
	return mux
}

// ApplyMiddlewares 应用插件中间件
func (m *manager) ApplyMiddlewares(handler http.Handler) http.Handler {
	plugins := m.registry.EnabledPlugins()
	
	for _, plug := range plugins {
		middlewares := plug.Middlewares()
		for _, middleware := range middlewares {
			// 应用中间件
			handler = middleware.Handle(context.Background(), handler)
			
			logger.Global().Debug("插件中间件",
				zap.String("plugin", plug.Name()),
				zap.String("middleware", middleware.Name()),
			)
		}
	}
	
	return handler
}

// ExecuteHooks 执行钩子
func (m *manager) ExecuteHooks(ctx context.Context, hookType HookType, data interface{}) error {
	plugins := m.registry.EnabledPlugins()
	
	hooks := make([]Hook, 0)
	
	// 收集所有钩子
	for _, plug := range plugins {
		hooks = append(hooks, plug.Hooks()...)
	}
	
	// 按优先级排序
	sortHooks(hooks)
	
	// 执行钩子
	for _, hook := range hooks {
		if hook.Type == hookType {
			if err := hook.Handler(ctx, data); err != nil {
				logger.Global().Error("钩子执行失败",
					zap.String("hook", hook.Name),
					zap.String("plugin", "unknown"),
					zap.String("error", err.Error()),
				)
				// 钩子错误不应影响其他钩子执行
				continue
			}
		}
	}
	
	return nil
}

// countEnabled 统计已启用插件数量
func (m *manager) countEnabled() int {
	count := 0
	for _, config := range m.config {
		if config.Enabled {
			count++
		}
	}
	return count
}

// sortHooks 按优先级排序钩子
func sortHooks(hooks []Hook) {
	// 简单的冒泡排序，优先级小的先执行
	for i := 0; i < len(hooks); i++ {
		for j := i + 1; j < len(hooks); j++ {
			if hooks[i].Priority > hooks[j].Priority {
				hooks[i], hooks[j] = hooks[j], hooks[i]
			}
		}
	}
}

// GetPluginConfig 获取插件配置
func (m *manager) GetPluginConfig(name string) (PluginConfig, bool) {
	config, exists := m.config[name]
	return config, exists
}

// UpdatePluginConfig 更新插件配置
func (m *manager) UpdatePluginConfig(name string, config PluginConfig) error {
	m.config[name] = config
	
	if config.Enabled {
		return m.registry.Enable(name)
	} else {
		return m.registry.Disable(name)
	}
}

// StopPlugin 停止单个插件
func (m *manager) StopPlugin(ctx context.Context, name string) error {
	plugin, exists := m.registry.Get(name)
	if !exists {
		return fmt.Errorf("插件 %s 不存在", name)
	}
	
	// 更新状态为停止中
	if reg, ok := m.registry.(*registry); ok {
		reg.setStatus(name, PluginStatusStopping)
	}
	
	// 停止插件
	if err := plugin.Stop(ctx); err != nil {
		return fmt.Errorf("停止插件失败: %w", err)
	}
	
	// 更新状态为已启用
	if reg, ok := m.registry.(*registry); ok {
		reg.setStatus(name, PluginStatusEnabled)
	}
	
	logger.Global().Info("插件停止成功", zap.String("plugin", name))
	
	return nil
}

// RestartPlugin 重启插件
func (m *manager) RestartPlugin(ctx context.Context, name string) error {
	// 先停止
	if err := m.StopPlugin(ctx, name); err != nil {
		return fmt.Errorf("停止插件失败: %w", err)
	}
	
	// 获取插件
	plugin, exists := m.registry.Get(name)
	if !exists {
		return fmt.Errorf("插件 %s 不存在", name)
	}
	
	// 获取配置
	config := map[string]interface{}{}
	if pluginConfig, exists := m.config[name]; exists {
		config = pluginConfig.Config
	}
	
	// 重新初始化
	if err := plugin.Init(ctx, config); err != nil {
		return fmt.Errorf("插件初始化失败: %w", err)
	}
	
	// 启动插件
	if err := plugin.Start(ctx); err != nil {
		return fmt.Errorf("插件启动失败: %w", err)
	}
	
	// 更新状态为运行中
	if reg, ok := m.registry.(*registry); ok {
		reg.setStatus(name, PluginStatusRunning)
	}
	
	logger.Global().Info("插件重启成功", zap.String("plugin", name))
	
	return nil
}