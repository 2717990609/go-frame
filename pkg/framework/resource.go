// Package framework 资源注册中心，供基础设施插件注册和获取共享资源
package framework

import (
	"sync"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

var (
	mu     sync.RWMutex
	globalDB   *gorm.DB
	globalRdb  *redis.Client
)

// RegisterDB 注册数据库连接（由 MySQL 插件调用）
func RegisterDB(db *gorm.DB) {
	mu.Lock()
	defer mu.Unlock()
	globalDB = db
}

// GetDB 获取已注册的数据库连接
func GetDB() *gorm.DB {
	mu.RLock()
	defer mu.RUnlock()
	return globalDB
}

// UnregisterDB 注销数据库连接（插件 Stop 时调用）
func UnregisterDB() {
	mu.Lock()
	defer mu.Unlock()
	globalDB = nil
}

// RegisterRedis 注册 Redis 客户端（由 Redis 插件调用）
func RegisterRedis(rdb *redis.Client) {
	mu.Lock()
	defer mu.Unlock()
	globalRdb = rdb
}

// GetRedis 获取已注册的 Redis 客户端
func GetRedis() *redis.Client {
	mu.RLock()
	defer mu.RUnlock()
	return globalRdb
}

// UnregisterRedis 注销 Redis 客户端（插件 Stop 时调用）
func UnregisterRedis() {
	mu.Lock()
	defer mu.Unlock()
	globalRdb = nil
}
