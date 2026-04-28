// Package mysql MySQL 数据库插件，以插件形式启动数据库连接
package mysql

import (
	"context"

	"go-backend-framework/pkg/framework"
	"go-backend-framework/pkg/logger"
	"go-backend-framework/pkg/plugin"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Plugin MySQL 数据库插件
type Plugin struct {
	plugin.BasePlugin
	db     *gorm.DB
	config Config
}

// Config MySQL 插件配置
type Config struct {
	Host          string `yaml:"host"`
	Port          int    `yaml:"port"`
	User          string `yaml:"user"`
	Password      string `yaml:"password"`
	Database      string `yaml:"database"`
	Charset       string `yaml:"charset"`
	Collation     string `yaml:"collation"`
	MaxOpenConns  int    `yaml:"max_open_conns"`
	MaxIdleConns  int    `yaml:"max_idle_conns"`
	EnableSQLLog  bool   `yaml:"enable_sql_log"`
}

// NewMySQLPlugin 创建 MySQL 插件
func NewMySQLPlugin() *Plugin {
	return &Plugin{
		BasePlugin: *plugin.NewBasePlugin("mysql", "1.0.0", "MySQL 数据库连接插件", "Framework Team"),
		config: Config{
			Host:         "localhost",
			Port:         3306,
			User:         "root",
			Password:     "",
			Database:     "default_db",
			Charset:      "utf8mb4",
			Collation:    "utf8mb4_general_ci",
			MaxOpenConns: 10,
			MaxIdleConns: 5,
		},
	}
}

// Init 解析配置
func (p *Plugin) Init(ctx context.Context, config map[string]interface{}) error {
	if v, ok := config["host"].(string); ok {
		p.config.Host = v
	}
	if v := getInt(config, "port"); v > 0 {
		p.config.Port = v
	}
	if v, ok := config["user"].(string); ok {
		p.config.User = v
	}
	if v, ok := config["password"].(string); ok {
		p.config.Password = v
	}
	if v, ok := config["database"].(string); ok {
		p.config.Database = v
	}
	if v, ok := config["charset"].(string); ok {
		p.config.Charset = v
	}
	if v, ok := config["collation"].(string); ok {
		p.config.Collation = v
	}
	if v := getInt(config, "max_open_conns"); v > 0 {
		p.config.MaxOpenConns = v
	}
	if v := getInt(config, "max_idle_conns"); v > 0 {
		p.config.MaxIdleConns = v
	}
	if v := getBool(config, "enable_sql_log"); v {
		p.config.EnableSQLLog = true
	}
	return nil
}

func getBool(m map[string]interface{}, key string) bool {
	v, ok := m[key]
	if !ok || v == nil {
		return false
	}
	switch b := v.(type) {
	case bool:
		return b
	case string:
		return b == "true" || b == "1"
	case int:
		return b != 0
	case int64:
		return b != 0
	case float64:
		return b != 0
	default:
		return false
	}
}

func getInt(m map[string]interface{}, key string) int {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}

// Start 建立数据库连接并注册到框架
func (p *Plugin) Start(ctx context.Context) error {
	db, err := framework.NewMySQL(framework.MySQLConfig{
		Host:         p.config.Host,
		Port:         p.config.Port,
		User:         p.config.User,
		Password:     p.config.Password,
		Database:     p.config.Database,
		Charset:      p.config.Charset,
		Collation:    p.config.Collation,
		MaxOpenConns: p.config.MaxOpenConns,
		MaxIdleConns: p.config.MaxIdleConns,
		EnableSQLLog: p.config.EnableSQLLog,
	})
	if err != nil {
		return err
	}
	p.db = db
	framework.RegisterDB(db)
	logger.Global().Info("MySQL 插件启动成功", zap.String("database", p.config.Database))
	return nil
}

// Stop 关闭连接并注销
func (p *Plugin) Stop(ctx context.Context) error {
	framework.UnregisterDB()
	if p.db != nil {
		sqlDB, err := p.db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
		p.db = nil
	}
	logger.Global().Info("MySQL 插件已停止")
	return nil
}
