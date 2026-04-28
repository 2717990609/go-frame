// Package config 配置加载兼容性测试
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_MinimalYAML(t *testing.T) {
	// 创建最小有效配置
	tmp := t.TempDir()
	yamlPath := filepath.Join(tmp, "config.yaml")
	content := `
server:
  port: 8080
  read_timeout: 30
  write_timeout: 30
  timeout: 30
mysql:
  host: localhost
  port: 3306
  user: root
  password: ""
  database: testdb
redis:
  addr: localhost:6379
  password: ""
  db: 0
log:
  output: ["stdout"]
`
	if err := os.WriteFile(yamlPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(yamlPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("server.port=%d, want 8080", cfg.Server.Port)
	}
	if cfg.MySQL.Database != "testdb" {
		t.Errorf("mysql.database=%s, want testdb", cfg.MySQL.Database)
	}
	if cfg.Redis.Addr != "localhost:6379" {
		t.Errorf("redis.addr=%s, want localhost:6379", cfg.Redis.Addr)
	}
}

func TestLoad_EnvVarExpansion(t *testing.T) {
	os.Setenv("COMPAT_TEST_PORT", "9999")
	os.Setenv("COMPAT_TEST_DB", "env_db")
	defer func() {
		os.Unsetenv("COMPAT_TEST_PORT")
		os.Unsetenv("COMPAT_TEST_DB")
	}()

	tmp := t.TempDir()
	yamlPath := filepath.Join(tmp, "config.yaml")
	content := `
server:
  port: ${COMPAT_TEST_PORT:8080}
mysql:
  host: localhost
  database: ${COMPAT_TEST_DB:default_db}
redis:
  addr: localhost:6379
log:
  output: ["stdout"]
`
	if err := os.WriteFile(yamlPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(yamlPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Server.Port != 9999 {
		t.Errorf("server.port=%d (env), want 9999", cfg.Server.Port)
	}
	if cfg.MySQL.Database != "env_db" {
		t.Errorf("mysql.database=%s (env), want env_db", cfg.MySQL.Database)
	}
}

func TestValidate(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: 8080},
		MySQL:  MySQLConfig{Host: "localhost", Database: "db"},
		Redis:  RedisConfig{Addr: "localhost:6379"},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("valid config should pass: %v", err)
	}
}

func TestValidate_Failures(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want string
	}{
		{"port zero", &Config{Server: ServerConfig{Port: 0}, MySQL: MySQLConfig{Host: "h", Database: "d"}, Redis: RedisConfig{Addr: "a"}}, "server.port"},
		{"mysql host empty", &Config{Server: ServerConfig{Port: 1}, MySQL: MySQLConfig{Host: "", Database: "d"}, Redis: RedisConfig{Addr: "a"}}, "mysql"},
		{"mysql database empty", &Config{Server: ServerConfig{Port: 1}, MySQL: MySQLConfig{Host: "h", Database: ""}, Redis: RedisConfig{Addr: "a"}}, "mysql"},
		{"redis addr empty", &Config{Server: ServerConfig{Port: 1}, MySQL: MySQLConfig{Host: "h", Database: "d"}, Redis: RedisConfig{Addr: ""}}, "redis"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cfg.Validate(); err == nil {
				t.Error("expected validation error")
			}
		})
	}
}
