// Package config 配置管理，支持多环境（规范第六章）
package config

import (
	"fmt"
	"os"
	"strings"

	pkgconfig "go-backend-framework/pkg/config"
	"go-backend-framework/pkg/logger"

	"gopkg.in/yaml.v3"
)

// Config 应用总配置
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	MySQL     MySQLConfig    `yaml:"mysql"`
	Redis     RedisConfig    `yaml:"redis"`
	Log       logger.Config  `yaml:"log"`
	JWT       JWTConfig      `yaml:"jwt"`
	Signature SignatureConfig `yaml:"signature"`
	CORS      CorsConfig     `yaml:"cors"`
}

// CorsConfig CORS 配置，支持 origins/methods/headers 等
type CorsConfig struct {
	Origins     string `yaml:"origins"`
	Methods     string `yaml:"methods"`
	Headers     string `yaml:"headers"`
	Credentials string `yaml:"credentials"`
}

// OriginsSlice 返回 origins 解析后的 []string，供 middleware.CORS 使用
func (c CorsConfig) OriginsSlice() []string {
	if c.Origins == "" {
		return []string{"*"}
	}
	parts := strings.Split(c.Origins, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// SignatureConfig 请求验签配置
type SignatureConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Key        string `yaml:"key"`
	TimeWindow int    `yaml:"time_window"` // 秒，默认 300
	NonceTTL   int    `yaml:"nonce_ttl"`   // 秒，默认 10
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	Port         int    `yaml:"port"`
	ReadTimeout  int    `yaml:"read_timeout"`  // 秒
	WriteTimeout int    `yaml:"write_timeout"` // 秒
	Timeout      int    `yaml:"timeout"`       // 请求超时秒
}

// MySQLConfig 数据库配置
type MySQLConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	Charset         string `yaml:"charset"`    // 字符集，默认 utf8mb4
	Collation       string `yaml:"collation"`  // 排序规则，默认 utf8mb4_general_ci
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// JWTConfig JWT 配置（密钥通过环境变量注入）
type JWTConfig struct {
	Secret     string `yaml:"secret"`
	ExpireHour int    `yaml:"expire_hour"`
}

// Load 加载配置文件，支持 ${VAR:default} 环境变量注入
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置失败: %w", err)
	}
	// 1. 先解析为 map，保留 ${VAR:default} 为字符串
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	// 2. 展开环境变量
	engine := pkgconfig.NewEngine()
	if err := engine.Load(raw); err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}
	// 3. 从 Engine 构建 Config（避免二次 unmarshal 类型问题）
	cfg := buildConfigFromEngine(engine)
	return cfg, nil
}

// buildConfigFromEngine 从配置引擎构建 Config 结构
func buildConfigFromEngine(e *pkgconfig.Engine) *Config {
	return &Config{
		Server: ServerConfig{
			Port:         e.GetInt("server.port", 8080),
			ReadTimeout:  e.GetInt("server.read_timeout", 30),
			WriteTimeout: e.GetInt("server.write_timeout", 30),
			Timeout:      e.GetInt("server.timeout", 30),
		},
		MySQL: MySQLConfig{
			Host:         e.GetString("mysql.host", "localhost"),
			Port:         e.GetInt("mysql.port", 3306),
			User:         e.GetString("mysql.user", "root"),
			Password:     e.GetString("mysql.password", ""),
			Database:     e.GetString("mysql.database", "default_db"),
			Charset:      e.GetString("mysql.charset", "utf8mb4"),
			Collation:    e.GetString("mysql.collation", "utf8mb4_general_ci"),
			MaxOpenConns: e.GetInt("mysql.max_open_conns", 10),
			MaxIdleConns: e.GetInt("mysql.max_idle_conns", 5),
		},
		Redis: RedisConfig{
			Addr:     e.GetString("redis.addr", "localhost:6379"),
			Password: e.GetString("redis.password", ""),
			DB:       e.GetInt("redis.db", 0),
		},
		Log: logger.Config{
			Output:          e.GetStringSlice("log.output"),
			FilePath:        e.GetString("log.file_path", ""),
			MaxSize:         e.GetInt("log.max_size", 100),
			MaxBackups:      e.GetInt("log.max_backups", 5),
			MaxAge:          e.GetInt("log.max_age", 7),
			Compress:        e.GetBool("log.compress", false),
			EnableSQLLog:    e.GetBool("log.enable_sql_log", false),
			SensitiveFields: e.GetStringSlice("log.sensitive_fields"),
		},
		JWT: JWTConfig{
			Secret:     e.GetString("jwt.secret", "default-secret"),
			ExpireHour: e.GetInt("jwt.expire_hour", 168),
		},
		Signature: SignatureConfig{
			Enabled:    e.GetBool("signature.enabled", false),
			Key:        e.GetString("signature.key", "default-key"),
			TimeWindow: e.GetInt("signature.time_window", 300),
			NonceTTL:   e.GetInt("signature.nonce_ttl", 10),
		},
		CORS: CorsConfig{
			Origins:     e.GetString("cors.origins", "*"),
			Methods:     e.GetString("cors.methods", "*"),
			Headers:     e.GetString("cors.headers", "*"),
			Credentials: e.GetString("cors.credentials", "true"),
		},
	}
}

// Validate 启动时校验必填配置
func (c *Config) Validate() error {
	if c.Server.Port <= 0 {
		return fmt.Errorf("server.port 必填")
	}
	if c.MySQL.Host == "" || c.MySQL.Database == "" {
		return fmt.Errorf("mysql.host 和 mysql.database 必填")
	}
	if c.Redis.Addr == "" {
		return fmt.Errorf("redis.addr 必填")
	}
	if c.Log.FilePath == "" && len(c.Log.Output) == 0 {
		c.Log.FilePath = "" // 默认 stdout
	}
	return nil
}
