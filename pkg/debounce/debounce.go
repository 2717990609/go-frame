// Package debounce 防抖通用工具，防止重复提交（规范 4.1）
package debounce

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// Apply 防抖检查，key 通过则设置 ttl，返回 true；已存在则返回 false
// 典型用法：用户提现、下单等操作前调用
func Apply(ctx context.Context, rdb *redis.Client, key string, ttl time.Duration) (bool, error) {
	ok, err := rdb.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("防抖检查失败: %w", err)
	}
	if !ok {
		return false, errors.New("操作进行中，请勿重复提交")
	}
	return true, nil
}
