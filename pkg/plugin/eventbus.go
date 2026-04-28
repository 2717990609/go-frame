// Package plugin 事件总线实现
package plugin

import (
	"sync"
)

// eventBus 事件总线实现
type eventBus struct {
	handlers map[EventType][]EventHandler
	mu       sync.RWMutex
}

// NewEventBus 创建事件总线
func NewEventBus() EventBus {
	return &eventBus{
		handlers: make(map[EventType][]EventHandler),
	}
}

// Subscribe 订阅事件
func (e *eventBus) Subscribe(eventType EventType, handler EventHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	e.handlers[eventType] = append(e.handlers[eventType], handler)
}

// Unsubscribe 取消订阅
func (e *eventBus) Unsubscribe(eventType EventType, handler EventHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	handlers := e.handlers[eventType]
	for i, h := range handlers {
		// 简单比较函数指针
		if &h == &handler {
			e.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// Publish 发布事件
func (e *eventBus) Publish(event Event) {
	e.mu.RLock()
	handlers := make([]EventHandler, len(e.handlers[event.Type]))
	copy(handlers, e.handlers[event.Type])
	e.mu.RUnlock()
	
	// 异步通知，避免阻塞
	for _, handler := range handlers {
		go func(h EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					// 记录事件处理器异常，但不应影响其他处理器
					// logger.Error("事件处理器异常", zap.Any("panic", r))
				}
			}()
			h(event)
		}(handler)
	}
}