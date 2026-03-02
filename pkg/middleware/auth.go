package middleware

import (
	"net/http"
	"strings"

	"fire-mirage/pkg/response"

	"github.com/gin-gonic/gin"
)

const userIDCtxKey = "user_id"

// Auth JWT/Token 校验，解析 user_id 注入 context（中间件链第 5 位）
// TokenFunc 由业务层实现，返回 userID 和是否有效
type TokenFunc func(token string) (userID int64, valid bool)

// Auth 创建认证中间件
func Auth(validateToken TokenFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, response.Error(response.CodePermissionError, "未登录或登录已过期"))
			c.Abort()
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, response.Error(response.CodePermissionError, "认证格式错误"))
			c.Abort()
			return
		}
		userID, valid := validateToken(parts[1])
		if !valid {
			c.JSON(http.StatusUnauthorized, response.Error(response.CodePermissionError, "登录已过期，请重新登录"))
			c.Abort()
			return
		}
		c.Set(userIDCtxKey, userID)
		c.Next()
	}
}

// GetUserID 从 context 获取当前用户 ID
func GetUserID(c *gin.Context) (int64, bool) {
	v, exists := c.Get(userIDCtxKey)
	if !exists {
		return 0, false
	}
	uid, ok := v.(int64)
	return uid, ok
}
