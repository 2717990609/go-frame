// Package framework 框架基础设施：数据库、Redis 等
package framework

import (
	"fmt"
	"time"

	sqldriver "github.com/go-sql-driver/mysql"
	"go-backend-framework/pkg/database/sqllogger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySQLConfig 简化配置
type MySQLConfig struct {
	Host          string
	Port          int
	User          string
	Password      string
	Database      string
	Charset       string // 字符集，默认 utf8mb4
	Collation     string // 排序规则，默认 utf8mb4_general_ci
	MaxOpenConns  int
	MaxIdleConns  int
	EnableSQLLog  bool   // 是否打印 SQL 日志（含 trace_id）
}

// NewMySQL 创建 MySQL 连接
func NewMySQL(cfg MySQLConfig) (*gorm.DB, error) {
	// 使用 mysql.Config 构建 DSN，自动处理密码中的 @、: 等特殊字符
	dsnCfg := sqldriver.Config{
		User:                 cfg.User,
		Passwd:               cfg.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		DBName:               cfg.Database,
		Loc:                  time.Local,
		ParseTime:            true,
		Params:               charsetParams(cfg.Charset, cfg.Collation),
	}
	dsn := dsnCfg.FormatDSN()
	gormConfig := &gorm.Config{}
	if cfg.EnableSQLLog {
		gormConfig.Logger = sqllogger.New(true)
	}
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("连接 MySQL 失败: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	return db, nil
}

func charsetParams(charset, collation string) map[string]string {
	if charset == "" {
		charset = "utf8mb4"
	}
	if collation == "" {
		collation = "utf8mb4_general_ci"
	}
	return map[string]string{"charset": charset, "collation": collation}
}
