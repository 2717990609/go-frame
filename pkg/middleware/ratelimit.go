package middleware

import (
	"net/http"
	"sync"
	"time"

	"fire-mirage/pkg/response"

	"github.com/gin-gonic/gin"
)

// RateLimit 简易内存限流（生产建议用 Redis 实现）（中间件链第 6 位）
type rateLimiter struct {
	mu       sync.Mutex
	counts   map[string]*countEntry
	limit    int
	interval time.Duration
}

type countEntry struct {
	count int
	until time.Time
}

// NewRateLimiter 全局限流，limit 为 interval 时间内最大请求数
func NewRateLimiter(limit int, interval time.Duration) *rateLimiter {
	r := &rateLimiter{
		counts:   make(map[string]*countEntry),
		limit:    limit,
		interval: interval,
	}
	go r.cleanup()
	return r
}

func (r *rateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		r.mu.Lock()
		now := time.Now()
		for k, v := range r.counts {
			if now.After(v.until) {
				delete(r.counts, k)
			}
		}
		r.mu.Unlock()
	}
}

func (r *rateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	e, ok := r.counts[key]
	if !ok || now.After(e.until) {
		r.counts[key] = &countEntry{count: 1, until: now.Add(r.interval)}
		return true
	}
	if e.count >= r.limit {
		return false
	}
	e.count++
	return true
}

// RateLimit 限流中间件，按 IP 限流
func RateLimit(limiter *rateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "ip:" + c.ClientIP()
		if !limiter.Allow(key) {
			c.JSON(http.StatusTooManyRequests, response.Error(response.CodeServerError, "请求过于频繁，请稍后再试"))
			c.Abort()
			return
		}
		c.Next()
	}
}
