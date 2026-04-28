// Package plugin 插件注册表实现
package plugin

import (
	"fmt"
	"sync"
	"time"

	"go-backend-framework/pkg/logger"
	"go.uber.org/zap"
)

// registry 插件注册表实现
type registry struct {
	plugins map[string]Plugin
	status  map[string]PluginStatus
	mu      sync.RWMutex
	events  EventBus
}

// NewRegistry 创建插件注册表
func NewRegistry() Registry {
	return &registry{
		plugins: make(map[string]Plugin),
		status:  make(map[string]PluginStatus),
		events:  NewEventBus(),
	}
}

// Register 注册插件
func (r *registry) Register(plug Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	name := plug.Name()
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("插件 %s 已存在", name)
	}
	
	r.plugins[name] = plug
	r.status[name] = PluginStatusDisabled
	
	// 发布注册事件
	r.events.Publish(Event{
		Type:      EventPluginRegistered,
		Plugin:    name,
		Message:   "插件注册成功",
		Timestamp: time.Now().Unix(),
	})
	
	logger.Global().Info("插件注册",
		zap.String("plugin", name),
		zap.String("version", plug.Version()),
	)
	
	return nil
}

// Unregister 注销插件
func (r *registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("插件 %s 不存在", name)
	}
	
	delete(r.plugins, name)
	delete(r.status, name)
	
	// 发布注销事件
	r.events.Publish(Event{
		Type:      EventPluginUnregistered,
		Plugin:    name,
		Message:   "插件注销成功",
		Timestamp: time.Now().Unix(),
	})
	
	logger.Global().Info("插件注销", zap.String("plugin", name))
	
	return nil
}

// Get 获取插件
func (r *registry) Get(name string) (Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	plugin, exists := r.plugins[name]
	return plugin, exists
}

// List 获取所有插件
func (r *registry) List() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	plugins := make([]Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}
	
	return plugins
}

// Enable 启用插件
func (r *registry) Enable(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("插件 %s 不存在", name)
	}
	
	r.status[name] = PluginStatusEnabled
	
	// 发布启用事件
	r.events.Publish(Event{
		Type:      EventPluginEnabled,
		Plugin:    name,
		Message:   "插件启用成功",
		Timestamp: time.Now().Unix(),
	})
	
	logger.Global().Info("插件启用", zap.String("plugin", name))
	
	return nil
}

// Disable 禁用插件
func (r *registry) Disable(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("插件 %s 不存在", name)
	}
	
	r.status[name] = PluginStatusDisabled
	
	// 发布禁用事件
	r.events.Publish(Event{
		Type:      EventPluginDisabled,
		Plugin:    name,
		Message:   "插件禁用成功",
		Timestamp: time.Now().Unix(),
	})
	
	logger.Global().Info("插件禁用", zap.String("plugin", name))
	
	return nil
}

// Status 获取插件状态
func (r *registry) Status(name string) PluginStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	status, exists := r.status[name]
	if !exists {
		return PluginStatusUnknown
	}
	
	return status
}

// setStatus 设置插件状态（内部方法）
func (r *registry) setStatus(name string, status PluginStatus) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.status[name] = status
}

// ListByStatus 根据状态列出插件
func (r *registry) ListByStatus(status PluginStatus) []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	plugins := make([]Plugin, 0)
	for name, plugin := range r.plugins {
		if r.status[name] == status {
			plugins = append(plugins, plugin)
		}
	}
	
	return plugins
}

// EnabledPlugins 获取已启用的插件
func (r *registry) EnabledPlugins() []Plugin {
	return r.ListByStatus(PluginStatusEnabled)
}

// DisabledPlugins 获取已禁用的插件
func (r *registry) DisabledPlugins() []Plugin {
	return r.ListByStatus(PluginStatusDisabled)
}

// Count 获取插件数量
func (r *registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return len(r.plugins)
}

// Exists 检查插件是否存在
func (r *registry) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	_, exists := r.plugins[name]
	return exists
}

// GetMetadata 获取插件元数据
func (r *registry) GetMetadata(name string) (*PluginMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	plugin, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("插件 %s 不存在", name)
	}
	
	return &PluginMetadata{
		Name:        plugin.Name(),
		Version:     plugin.Version(),
		Description: plugin.Description(),
		Author:      plugin.Author(),
		Tags:        []string{}, // 需要从插件中获取
		Dependencies: []string{}, // 需要从插件中获取
		Config:      map[string]interface{}{}, // 需要从插件中获取
	}, nil
}

// UpdateStatus 更新插件状态
func (r *registry) UpdateStatus(name string, status PluginStatus) error {
	if !r.Exists(name) {
		return fmt.Errorf("插件 %s 不存在", name)
	}
	
	r.setStatus(name, status)
	return nil
}

// SubscribeEvent 订阅插件事件
func (r *registry) SubscribeEvent(eventType EventType, handler EventHandler) {
	r.events.Subscribe(eventType, handler)
}

// PublishEvent 发布插件事件
func (r *registry) PublishEvent(eventType EventType, plugin, message string, data interface{}) {
	r.events.Publish(Event{
		Type:      eventType,
		Plugin:    plugin,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// GetEventBus 获取事件总线
func (r *registry) GetEventBus() EventBus {
	return r.events
}