package middleware

import (
	"time"

	"fire-mirage/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger 记录请求路径、耗时、状态码（中间件链第 3 位）
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		method := c.Request.Method
		c.Next()
		latency := time.Since(start)
		status := c.Writer.Status()
		ctx := c.Request.Context()
		logger.C(ctx).Info("request",
			zap.String("path", path),
			zap.String("method", method),
			zap.String("ip", clientIP),
			zap.Int("status", status),
			zap.Duration("duration", latency),
		)
	}
}
