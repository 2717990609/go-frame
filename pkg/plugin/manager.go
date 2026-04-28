// Package plugin 插件管理器实现
package plugin

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go-backend-framework/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// manager 插件管理器实现
type manager struct {
	registry Registry
	plugins  []Plugin
	config   map[string]PluginConfig
	metrics  map[string]PluginMetrics
}

// NewManager 创建插件管理器
func NewManager() Manager {
	return &manager{
		registry: NewRegistry(),
		plugins:  make([]Plugin, 0),
		config:   make(map[string]PluginConfig),
		metrics:  make(map[string]PluginMetrics),
	}
}

// Register 注册插件
func (m *manager) Register(plugin Plugin) error {
	start := time.Now()
	if err := m.registry.Register(plugin); err != nil {
		return err
	}

	name := plugin.Name()
	metric := m.metrics[name]
	metric.LoadDurationMs = time.Since(start).Milliseconds()
	m.metrics[name] = metric
	return nil
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

	// 先进行配置校验与依赖检查
	for pluginName, pluginConfig := range config {
		plug, exists := m.registry.Get(pluginName)
		if !exists {
			return fmt.Errorf("插件 %s 未注册", pluginName)
		}
		// 禁用插件只做注册存在性检查，不做配置/依赖强校验
		if !pluginConfig.Enabled {
			continue
		}
		if err := m.validateConfig(plug, pluginConfig.Config); err != nil {
			return fmt.Errorf("插件 %s 配置校验失败: %w", pluginName, err)
		}
		if err := m.validateDependencies(pluginName, pluginConfig); err != nil {
			return err
		}
	}

	// 根据配置启用/禁用插件
	for pluginName, pluginConfig := range config {
		if pluginConfig.Enabled {
			if err := m.registry.Enable(pluginName); err != nil {
				logger.Global().Error("启用插件失败",
					zap.String("plugin", pluginName),
					zap.String("error", err.Error()),
				)
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
		start := time.Now()

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
			m.markFailure(pluginName)

			// 更新状态为错误
			if reg, ok := m.registry.(*registry); ok {
				reg.setStatus(pluginName, PluginStatusError)
				reg.PublishEvent(EventPluginError, pluginName, "插件初始化失败", map[string]interface{}{
					"stage": "init",
					"error": err.Error(),
				})
			}
			if m.isCriticalPlugin(pluginName, plug) {
				return fmt.Errorf("关键插件 %s 初始化失败: %w", pluginName, err)
			}
			continue
		}

		// 更新状态为已初始化
		if reg, ok := m.registry.(*registry); ok {
			reg.setStatus(pluginName, PluginStatusInitialized)
		}
		m.updateMetrics(pluginName, func(metric *PluginMetrics) {
			metric.InitDurationMs = time.Since(start).Milliseconds()
		})

		logger.Global().Info("插件初始化成功",
			zap.String("plugin", pluginName),
		)
	}

	return nil
}

// Start 启动所有插件
func (m *manager) Start(ctx context.Context) error {
	plugins := m.registry.List()

	for _, plug := range plugins {
		pluginName := plug.Name()
		status := m.registry.Status(pluginName)
		if status != PluginStatusEnabled && status != PluginStatusInitialized {
			continue
		}
		start := time.Now()

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
			m.markFailure(pluginName)

			// 更新状态为错误
			if reg, ok := m.registry.(*registry); ok {
				reg.setStatus(pluginName, PluginStatusError)
				reg.PublishEvent(EventPluginError, pluginName, "插件启动失败", map[string]interface{}{
					"stage": "start",
					"error": err.Error(),
				})
			}
			if m.isCriticalPlugin(pluginName, plug) {
				return fmt.Errorf("关键插件 %s 启动失败: %w", pluginName, err)
			}
			continue
		}

		// 更新状态为运行中
		if reg, ok := m.registry.(*registry); ok {
			reg.setStatus(pluginName, PluginStatusRunning)
		}
		m.updateMetrics(pluginName, func(metric *PluginMetrics) {
			metric.StartDurationMs = time.Since(start).Milliseconds()
		})

		// 发布启动事件
		if reg, ok := m.registry.(*registry); ok {
			reg.PublishEvent(EventPluginStarted, pluginName, "插件启动成功", m.metrics[pluginName])
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
		start := time.Now()
		if err := plug.Stop(ctx); err != nil {
			logger.Global().Error("插件停止失败",
				zap.String("plugin", pluginName),
				zap.String("error", err.Error()),
			)
			m.markFailure(pluginName)
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
		m.updateMetrics(pluginName, func(metric *PluginMetrics) {
			metric.StopDurationMs = time.Since(start).Milliseconds()
		})

		logger.Global().Info("插件停止成功",
			zap.String("plugin", pluginName),
		)
	}

	return nil
}

// ApplyRoutes 应用插件路由
func (m *manager) ApplyRoutes(mux http.Handler) http.Handler {
	engine, ok := mux.(*gin.Engine)
	if !ok {
		logger.Global().Warn("插件路由接入跳过：仅支持 gin.Engine")
		return mux
	}

	plugins := m.registry.EnabledPlugins()

	for _, plug := range plugins {
		if m.registry.Status(plug.Name()) != PluginStatusRunning {
			continue
		}
		routes := plug.Routes()
		for _, route := range routes {
			method := strings.ToUpper(route.Method)
			handler := route.Handler
			if handler == nil {
				continue
			}
			engine.Handle(method, route.Path, gin.WrapF(handler))
			logger.Global().Debug("插件路由",
				zap.String("plugin", plug.Name()),
				zap.String("method", method),
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
		if m.registry.Status(plug.Name()) != PluginStatusRunning {
			continue
		}
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
	plug, exists := m.registry.Get(name)
	if !exists {
		return fmt.Errorf("插件 %s 不存在", name)
	}
	if err := m.validateConfig(plug, config.Config); err != nil {
		return fmt.Errorf("插件 %s 配置校验失败: %w", name, err)
	}
	if err := m.validateDependencies(name, config); err != nil {
		return err
	}
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
	start := time.Now()
	if err := plugin.Stop(ctx); err != nil {
		return fmt.Errorf("停止插件失败: %w", err)
	}

	// 更新状态为已启用
	if reg, ok := m.registry.(*registry); ok {
		reg.setStatus(name, PluginStatusEnabled)
	}
	m.updateMetrics(name, func(metric *PluginMetrics) {
		metric.StopDurationMs = time.Since(start).Milliseconds()
	})

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

// GetMetrics 获取插件指标快照
func (m *manager) GetMetrics() map[string]PluginMetrics {
	snapshot := make(map[string]PluginMetrics, len(m.metrics))
	for name, metric := range m.metrics {
		snapshot[name] = metric
	}
	return snapshot
}

func (m *manager) validateConfig(plug Plugin, cfg map[string]interface{}) error {
	if cfg == nil {
		cfg = map[string]interface{}{}
	}
	if aware, ok := plug.(ConfigAwarePlugin); ok {
		return aware.ValidateConfig(cfg)
	}
	return nil
}

func (m *manager) validateDependencies(pluginName string, cfg PluginConfig) error {
	deps := append([]string{}, cfg.Dependencies...)
	if plug, exists := m.registry.Get(pluginName); exists {
		if aware, ok := plug.(DependencyAwarePlugin); ok {
			deps = append(deps, aware.Dependencies()...)
		}
	}
	seen := map[string]struct{}{}
	for _, dep := range deps {
		if dep == "" {
			continue
		}
		if _, duplicated := seen[dep]; duplicated {
			continue
		}
		seen[dep] = struct{}{}
		depCfg, exists := m.config[dep]
		if !exists {
			return fmt.Errorf("插件 %s 依赖 %s 缺少配置", pluginName, dep)
		}
		if _, exists := m.registry.Get(dep); !exists {
			return fmt.Errorf("插件 %s 依赖 %s 未注册", pluginName, dep)
		}
		if !depCfg.Enabled {
			return fmt.Errorf("插件 %s 依赖 %s 未启用", pluginName, dep)
		}
	}
	m.updateMetrics(pluginName, func(metric *PluginMetrics) {
		metric.DependencyCount = len(seen)
	})
	return nil
}

func (m *manager) isCriticalPlugin(name string, plug Plugin) bool {
	if cfg, ok := m.config[name]; ok && cfg.Critical {
		return true
	}
	if critical, ok := plug.(CriticalPlugin); ok {
		return critical.Critical()
	}
	return false
}

func (m *manager) markFailure(name string) {
	m.updateMetrics(name, func(metric *PluginMetrics) {
		metric.FailureCount++
	})
}

func (m *manager) updateMetrics(name string, updater func(metric *PluginMetrics)) {
	metric := m.metrics[name]
	updater(&metric)
	m.metrics[name] = metric
}
