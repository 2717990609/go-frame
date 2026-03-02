// Package service 业务逻辑层
package service

import (
	"context"

	"fire-mirage/internal/dto"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// HealthService 健康检查服务
type HealthService struct {
	db   *gorm.DB
	rdb  *redis.Client
}

// NewHealthService 创建
func NewHealthService(db *gorm.DB, rdb *redis.Client) *HealthService {
	return &HealthService{db: db, rdb: rdb}
}

// Health 存活探测，无依赖检查
func (s *HealthService) Health(ctx context.Context) dto.HealthResponse {
	return dto.HealthResponse{Status: "ok"}
}

// Ready 就绪探测，检查 MySQL、Redis 连通性
func (s *HealthService) Ready(ctx context.Context) dto.ReadyResponse {
	resp := dto.ReadyResponse{
		Ready:   true,
		MySQL:   false,
		Redis:   false,
		Details: make(map[string]string),
	}
	if s.db != nil {
		if err := s.db.WithContext(ctx).Exec("SELECT 1").Error; err != nil {
			resp.Details["mysql"] = err.Error()
		} else {
			resp.MySQL = true
		}
	}
	if s.rdb != nil {
		if err := s.rdb.Ping(ctx).Err(); err != nil {
			resp.Details["redis"] = err.Error()
		} else {
			resp.Redis = true
		}
	}
	resp.Ready = resp.MySQL && resp.Redis
	return resp
}
