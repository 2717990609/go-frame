// Package compatibility API 兼容性测试（响应格式、路由结构）
package compatibility

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-backend-framework/pkg/response"

	"github.com/gin-gonic/gin"
)

func TestAPI_ResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, response.Success(map[string]string{"pong": "ok"}))
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status=%d, want 200", w.Code)
	}
	body := w.Body.Bytes()
	if len(body) == 0 {
		t.Fatal("empty body")
	}
	// 校验标准响应结构：code, msg, data, ts
	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	for _, key := range []string{"code", "msg", "data", "ts"} {
		if _, ok := m[key]; !ok {
			t.Errorf("response missing key %q", key)
		}
	}
	if code, _ := m["code"].(float64); code != float64(response.CodeSuccess) {
		t.Errorf("code=%v, want %d", code, response.CodeSuccess)
	}
}

func TestAPI_HealthPath(t *testing.T) {
	// 验证健康检查路径存在且返回 200（使用模拟 handler，不依赖真实 DB）
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, response.Success(map[string]string{"status": "ok"}))
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("/health status=%d, want 200", w.Code)
	}
}

