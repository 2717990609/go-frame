package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORS 跨域配置（中间件链第 7 位）
func CORS(allowedOrigins []string) gin.HandlerFunc {
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allow := "*"
		for _, o := range allowedOrigins {
			if o == "*" || o == origin {
				allow = o
				if o == "*" {
					allow = "*"
				}
				break
			}
		}
		c.Header("Access-Control-Allow-Origin", allow)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Trace-Id")
		c.Header("Access-Control-Max-Age", "86400")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
