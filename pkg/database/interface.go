// Package database 数据库提供者抽象接口
package database

import (
	"context"
	"time"
)

// Provider 数据库提供者接口
type Provider interface {
	// 基础连接操作
	Connect(ctx context.Context, config Config) error
	Close() error
	Ping(ctx context.Context) error
	
	// 基础CRUD操作
	Find(ctx context.Context, dest any, query Query) error
	FindOne(ctx context.Context, dest any, query Query) error
	Create(ctx context.Context, model any) error
	Update(ctx context.Context, model any, updates any) error
	Delete(ctx context.Context, models ...any) error
	
	// 事务支持
	Transaction(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error
	
	// 原生SQL支持
	Exec(ctx context.Context, query string, args ...any) (Result, error)
	Query(ctx context.Context, query string, args ...any) (Rows, error)
	
	// 特定能力接口
	Migrator() Migrator
	Builder() QueryBuilder
	Dialect() Dialect
}

// Tx 事务接口
type Tx interface {
	// 基础CRUD操作（与Provider相同签名）
	Find(ctx context.Context, dest any, query Query) error
	FindOne(ctx context.Context, dest any, query Query) error
	Create(ctx context.Context, model any) error
	Update(ctx context.Context, model any, updates any) error
	Delete(ctx context.Context, models ...any) error
	
	// 事务操作
	Commit() error
	Rollback() error
	
	// 原生SQL支持
	Exec(ctx context.Context, query string, args ...any) (Result, error)
	Query(ctx context.Context, query string, args ...any) (Rows, error)
}

// Query 查询条件
type Query struct {
	Where    Conditions    `json:"where"`
	Order    []Order       `json:"order"`
	Limit    *int          `json:"limit"`
	Offset   *int          `json:"offset"`
	Joins    []string      `json:"joins"`
	Select   []string      `json:"select"`
	GroupBy  []string      `json:"group_by"`
	Having   Conditions    `json:"having"`
}

// Conditions 查询条件
type Conditions []Condition

// Condition 单个条件
type Condition struct {
	Column   string      `json:"column"`
	Operator string      `json:"operator"` // =, !=, >, <, >=, <=, like, in, not_in, is_null, is_not_null
	Value    interface{} `json:"value"`
}

// Order 排序条件
type Order struct {
	Column string `json:"column"`
	Direction string `json:"direction"` // asc, desc
}

// Result 执行结果
type Result struct {
	RowsAffected int64
	LastInsertID int64
}

// Rows 查询结果集
type Rows interface {
	Close() error
	Next() bool
	Scan(dest ...any) error
	Columns() ([]string, error)
}

// Migrator 数据库迁移接口
type Migrator interface {
	CreateTable(ctx context.Context, model any) error
	DropTable(ctx context.Context, model any) error
	HasTable(ctx context.Context, tableName string) (bool, error)
	AddColumn(ctx context.Context, tableName, column string, fieldType interface{}) error
	DropColumn(ctx context.Context, tableName, column string) error
	AutoMigrate(ctx context.Context, models ...any) error
}

// QueryBuilder 查询构建器接口
type QueryBuilder interface {
	Select(columns ...string) QueryBuilder
	From(table string) QueryBuilder
	Where(conditions ...Condition) QueryBuilder
	OrderBy(column string, direction string) QueryBuilder
	Limit(limit int) QueryBuilder
	Offset(offset int) QueryBuilder
	Join(table string, on string) QueryBuilder
	GroupBy(columns ...string) QueryBuilder
	Having(conditions ...Condition) QueryBuilder
	Build() (string, []any)
}

// Dialect 数据库方言
type Dialect interface {
	// SQL语法差异处理
	Quote(identifier string) string
	Now() string
	// 数据类型映射
	ColumnType(fieldType string) string
	// 特定功能支持
	SupportsLastInsertID() bool
	SupportsReturningClause() bool
}

// Config 数据库配置
type Config struct {
	Driver      string        `json:"driver"`
	Host        string        `json:"host"`
	Port        int           `json:"port"`
	Database    string        `json:"database"`
	Username    string        `json:"username"`
	Password    string        `json:"password"`
	Charset     string        `json:"charset"`
	Collation   string        `json:"collation"`
	MaxOpenConns int          `json:"max_open_conns"`
	MaxIdleConns int          `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
	// 额外配置参数
	Params      map[string]interface{} `json:"params"`
}

// Registry 数据库提供者注册表
var Registry = make(map[string]func(Config) Provider)

// Register 注册数据库提供者
func Register(driver string, factory func(Config) Provider) {
	Registry[driver] = factory
}

// New 创建数据库提供者
func New(config Config) (Provider, error) {
	factory, exists := Registry[config.Driver]
	if !exists {
		return nil, ErrUnsupportedDriver
	}
	
	provider := factory(config)
	return provider, nil
}

// 错误定义
var (
	ErrUnsupportedDriver = fmt.Errorf("unsupported database driver")
	ErrConnectionFailed  = fmt.Errorf("database connection failed")
	ErrQueryFailed       = fmt.Errorf("query execution failed")
	ErrTransactionFailed = fmt.Errorf("transaction failed")
)

// 便利方法

// Q 快速构建Query条件
func Q() *QueryBuilder {
	return &QueryBuilder{}
}

// Where 构建WHERE条件
func Where(column, operator string, value interface{}) Condition {
	return Condition{Column: column, Operator: operator, Value: value}
}

// Eq 等于条件
func Eq(column string, value interface{}) Condition {
	return Condition{Column: column, Operator: "=", Value: value}
}

// Ne 不等于条件
func Ne(column string, value interface{}) Condition {
	return Condition{Column: column, Operator: "!=", Value: value}
}

// Gt 大于条件
func Gt(column string, value interface{}) Condition {
	return Condition{Column: column, Operator: ">", Value: value}
}

// Lt 小于条件
func Lt(column string, value interface{}) Condition {
	return Condition{Column: column, Operator: "<", Value: value}
}

// Like LIKE条件
func Like(column, pattern string) Condition {
	return Condition{Column: column, Operator: "like", Value: pattern}
}

// In IN条件
func In(column string, values ...interface{}) Condition {
	return Condition{Column: column, Operator: "in", Value: values}
}

// IsNull IS NULL条件
func IsNull(column string) Condition {
	return Condition{Column: column, Operator: "is_null", Value: nil}
}

// IsNotNULL IS NOT NULL条件
func IsNotNull(column string) Condition {
	return Condition{Column: column, Operator: "is_not_null", Value: nil}
}

// OrderBy 构建排序
func OrderBy(column, direction string) Order {
	return Order{Column: column, Direction: direction}
}

// Asc 升序排序
func Asc(column string) Order {
	return Order{Column: column, Direction: "asc"}
}

// Desc 降序排序
func Desc(column string) Order {
	return Order{Column: column, Direction: "desc"}
}