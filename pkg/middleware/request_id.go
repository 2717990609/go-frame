package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"go-backend-framework/pkg/logger"

	"github.com/gin-gonic/gin"
)

const traceIDHeader = "X-Trace-Id"
const traceIDCtxKey = "trace_id"

// RequestID 生成 trace_id，注入 context 与 response header（中间件链第 2 位）
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(traceIDHeader)
		if traceID == "" {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			traceID = "req-" + hex.EncodeToString(b)
		}
		c.Set(traceIDCtxKey, traceID)
		c.Header(traceIDHeader, traceID)
		ctx := logger.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// GetTraceID 从 gin.Context 获取 trace_id
func GetTraceID(c *gin.Context) string {
	return c.GetString(traceIDCtxKey)
}
