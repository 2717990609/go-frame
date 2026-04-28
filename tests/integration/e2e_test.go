// Package integration 端到端集成测试
// 需要 MySQL 和 Redis 可用时运行: go test -v ./tests/integration/...
// 跳过集成测试: go test -short ./...
package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-backend-framework/api"
	"go-backend-framework/config"
	"go-backend-framework/pkg/framework"
	"go-backend-framework/pkg/logger"
	"go-backend-framework/pkg/plugin"
	"go-backend-framework/pkg/plugin/builtin"

	"github.com/gin-gonic/gin"
)

func TestE2E_ServerStartAndHealth(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e in short mode")
	}
	// 1. 加载配置
	cfg, err := config.Load("config/config.dev.yaml")
	if err != nil {
		t.Skipf("config load failed (skip e2e): %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("config invalid: %v", err)
	}
	// 2. 初始化日志（静默）
	_ = logger.Init(cfg.Log)

	// 3. 启动插件（MySQL、Redis）
	pluginMgr := plugin.NewManager()
	ctx := context.Background()
	if err := builtin.LoadBuiltin(ctx, pluginMgr, cfg); err != nil {
		t.Skipf("plugins failed (skip e2e, need MySQL/Redis): %v", err)
	}
	db := framework.GetDB()
	rdb := framework.GetRedis()
	if db == nil || rdb == nil {
		t.Skip("MySQL or Redis not available, skip e2e")
	}
	defer pluginMgr.Stop(ctx)

	// 4. 组装路由
	engine, err := api.Setup(cfg, db, rdb)
	if err != nil {
		t.Fatalf("api setup: %v", err)
	}

	// 5. 使用 httptest 模拟请求
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name   string
		path   string
		status int
	}{
		{"health", "/health", http.StatusOK},
		{"ping", "/ping", http.StatusOK},
		{"ready", "/ready", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			if w.Code != tt.status {
				t.Errorf("%s: status=%d, want %d", tt.path, w.Code, tt.status)
			}
		})
	}
}

func TestE2E_RealHTTP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e in short mode")
	}
	cfg, err := config.Load("config/config.dev.yaml")
	if err != nil {
		t.Skipf("config: %v", err)
	}
	_ = logger.Init(cfg.Log)
	pluginMgr := plugin.NewManager()
	ctx := context.Background()
	if err := builtin.LoadBuiltin(ctx, pluginMgr, cfg); err != nil {
		t.Skipf("plugins: %v", err)
	}
	db := framework.GetDB()
	rdb := framework.GetRedis()
	if db == nil || rdb == nil {
		t.Skip("DB/Redis not available")
	}
	engine, err := api.Setup(cfg, db, rdb)
	if err != nil {
		t.Fatalf("api setup: %v", err)
	}
	// 使用 httptest.Server 模拟真实 HTTP 请求
	srv := httptest.NewServer(engine)
	defer srv.Close()
	defer pluginMgr.Stop(ctx)

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("http get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("health status=%d", resp.StatusCode)
	}
}
