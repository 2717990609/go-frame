// Package benchmark 性能基准测试（Phase 2 技术栈升级验收）
package benchmark

import (
	"os"
	"path/filepath"
	"testing"

	"go-backend-framework/config"
	pkgconfig "go-backend-framework/pkg/config"
)

// BenchmarkConfigEngine_Load 配置引擎加载性能
func BenchmarkConfigEngine_Load(b *testing.B) {
	data := map[string]interface{}{
		"server": map[string]interface{}{
			"port":         8080,
			"read_timeout": 30,
			"write_timeout": 30,
		},
		"mysql": map[string]interface{}{
			"host":     "localhost",
			"port":     3306,
			"database": "testdb",
		},
		"redis": map[string]interface{}{
			"addr": "localhost:6379",
		},
	}
	e := pkgconfig.NewEngine()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.Load(data)
	}
}

// BenchmarkConfigEngine_GetNested 嵌套配置访问性能
func BenchmarkConfigEngine_GetNested(b *testing.B) {
	e := pkgconfig.NewEngine()
	e.Load(map[string]interface{}{
		"server": map[string]interface{}{"port": 8080},
		"mysql":  map[string]interface{}{"database": "db"},
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.Get("mysql.database")
	}
}

// BenchmarkConfigEngine_ExpandEnv 环境变量展开性能
func BenchmarkConfigEngine_ExpandEnv(b *testing.B) {
	os.Setenv("BENCH_PORT", "8080")
	defer os.Unsetenv("BENCH_PORT")
	data := map[string]interface{}{
		"server": map[string]interface{}{
			"port": "${BENCH_PORT:8080}",
		},
	}
	e := pkgconfig.NewEngine()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.Load(data)
	}
}

// BenchmarkConfigLoad_Full 完整配置文件加载（需存在 config.dev.yaml）
func BenchmarkConfigLoad_Full(b *testing.B) {
	path := "config/config.dev.yaml"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = filepath.Join("..", "..", "config", "config.dev.yaml")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		b.Skip("config file not found")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = config.Load(path)
	}
}
