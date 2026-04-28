// Package plugin 插件系统接口定义
package plugin

import (
	"context"
	"net/http"
)

// Plugin 插件标准接口
type Plugin interface {
	// 插件元信息
	Name() string
	Version() string
	Description() string
	Author() string
	
	// 生命周期
	Init(ctx context.Context, config map[string]interface{}) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	
	// 功能接口
	Routes() []Route
	Middlewares() []Middleware
	Hooks() []Hook
}

// Route 路由定义
type Route struct {
	Method   string
	Path     string
	Handler  http.HandlerFunc
	// 可选的中间件
	Middlewares []Middleware
	// 路由元信息
	Description string
	Tags        []string
}

// Middleware 中间件定义
type Middleware interface {
	// 中间件元信息
	Name() string
	Description() string
	
	// 中间件执行
	Handle(ctx context.Context, next http.Handler) http.Handler
}

// Hook 钩子定义
type Hook struct {
	// 钩子类型
	Type HookType
	// 钩子名称
	Name string
	// 钩子处理器
	Handler HookHandler
	// 执行优先级
	Priority int
}

// HookType 钩子类型枚举
type HookType int

const (
	// HookBeforeRequest 请求前钩子
	HookBeforeRequest HookType = iota
	// HookAfterRequest 请求后钩子
	HookAfterRequest
	// HookBeforeResponse 响应前钩子
	HookBeforeResponse
	// HookAfterResponse 响应后钩子
	HookAfterResponse
	// HookOnError 错误钩子
	HookOnError
	// HookOnStartup 启动钩子
	HookOnStartup
	// HookOnShutdown 关闭钩子
	HookOnShutdown
)

// HookHandler 钩子处理器
type HookHandler func(ctx context.Context, data interface{}) error

// HookData 钩子数据
type HookData struct {
	Type  HookType
	Name  string
	Data  interface{}
	Error error
}

// PluginStatus 插件状态
type PluginStatus int

const (
	// PluginStatusUnknown 未知状态
	PluginStatusUnknown PluginStatus = iota
	// PluginStatusDisabled 已禁用
	PluginStatusDisabled
	// PluginStatusEnabled 已启用
	PluginStatusEnabled
	// PluginStatusStarting 启动中
	PluginStatusStarting
	// PluginStatusRunning 运行中
	PluginStatusRunning
	// PluginStatusStopping 停止中
	PluginStatusStopping
	// PluginStatusError 错误状态
	PluginStatusError
)

// PluginMetadata 插件元数据
type PluginMetadata struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Tags        []string          `json:"tags"`
	Dependencies []string         `json:"dependencies"`
	Config      map[string]interface{} `json:"config"`
}

// PluginConfig 插件配置
type PluginConfig struct {
	Enabled bool                     `yaml:"enabled"`
	Config  map[string]interface{}   `yaml:"config"`
}

// Registry 插件注册表接口
type Registry interface {
	// 注册插件
	Register(plugin Plugin) error
	// 注销插件
	Unregister(name string) error
	// 获取插件
	Get(name string) (Plugin, bool)
	// 获取所有插件
	List() []Plugin
	// 启用插件
	Enable(name string) error
	// 禁用插件
	Disable(name string) error
	// 获取插件状态
	Status(name string) PluginStatus
	// 获取已启用的插件
	EnabledPlugins() []Plugin
}

// Manager 插件管理器接口
type Manager interface {
	Registry
	
	// 加载插件配置
	LoadConfig(config map[string]PluginConfig) error
	// 初始化所有插件
	Initialize(ctx context.Context) error
	// 启动所有插件
	Start(ctx context.Context) error
	// 停止所有插件
	Stop(ctx context.Context) error
	
	// 集成到HTTP服务器
	ApplyRoutes(mux http.Handler) http.Handler
	ApplyMiddlewares(handler http.Handler) http.Handler
	
	// 执行钩子
	ExecuteHooks(ctx context.Context, hookType HookType, data interface{}) error
}

// Loader 插件加载器接口
type Loader interface {
	// 从目录加载插件
	LoadFromDir(dir string) error
	// 从文件加载插件
	LoadFromFile(file string) error
	// 加载内置插件
	LoadBuiltin() error
}

// Event 插件事件
type Event struct {
	Type      EventType `json:"type"`
	Plugin    string    `json:"plugin"`
	Message   string    `json:"message"`
	Data      interface{} `json:"data"`
	Timestamp int64     `json:"timestamp"`
	Error     string    `json:"error,omitempty"`
}

// EventType 事件类型
type EventType int

const (
	// EventPluginRegistered 插件注册事件
	EventPluginRegistered EventType = iota
	// EventPluginUnregistered 插件注销事件
	EventPluginUnregistered
	// EventPluginEnabled 插件启用事件
	EventPluginEnabled
	// EventPluginDisabled 插件禁用事件
	EventPluginDisabled
	// EventPluginStarted 插件启动事件
	EventPluginStarted
	// EventPluginStopped 插件停止事件
	EventPluginStopped
	// EventPluginError 插件错误事件
	EventPluginError
)

// EventHandler 事件处理器
type EventHandler func(event Event)

// EventBus 事件总线接口
type EventBus interface {
	// 订阅事件
	Subscribe(eventType EventType, handler EventHandler)
	// 发布事件
	Publish(event Event)
	// 取消订阅
	Unsubscribe(eventType EventType, handler EventHandler)
}