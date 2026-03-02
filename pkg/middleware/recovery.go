// Package middleware HTTP 中间件体系（规范第七章）
package middleware

import (
	"net/http"

	"fire-mirage/pkg/logger"
	"fire-mirage/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery panic 捕获中间件，返回 500，记录堆栈（中间件链第 1 位）
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				ctx := c.Request.Context()
				traceID := c.GetString("trace_id")
				if traceID != "" {
					ctx = logger.WithTraceID(ctx, traceID)
				}
				logger.C(ctx).Error("panic recovered",
					zap.Any("recover", r),
					zap.Stack("stack"),
					zap.String("path", c.Request.URL.Path),
				)
				c.JSON(http.StatusInternalServerError, response.Error(response.CodeServerError, "系统繁忙，请稍后重试"))
				c.Abort()
			}
		}()
		c.Next()
	}
}
