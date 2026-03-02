package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout 统一 context.WithTimeout（默认 30s）（中间件链第 4 位）
func Timeout(d time.Duration) func(*gin.Context) {
	if d <= 0 {
		d = 30 * time.Second
	}
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
