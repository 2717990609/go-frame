// Package config 配置管理，支持多环境（规范第六章）
package config

import (
	"fmt"
	"os"

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
	CORS      []string       `yaml:"cors"`
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

// Load 加载配置文件，支持环境变量覆盖
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置失败: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	// 环境变量覆盖敏感配置
	if s := os.Getenv("MYSQL_PASSWORD"); s != "" {
		cfg.MySQL.Password = s
	}
	if s := os.Getenv("REDIS_PASSWORD"); s != "" {
		cfg.Redis.Password = s
	}
	if s := os.Getenv("JWT_SECRET"); s != "" {
		cfg.JWT.Secret = s
	}
	if s := os.Getenv("SIGNATURE_KEY"); s != "" {
		cfg.Signature.Key = s
	}
	return &cfg, nil
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
