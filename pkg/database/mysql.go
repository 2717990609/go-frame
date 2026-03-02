// Package database MySQL数据库实现
package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go-backend-framework/pkg/logger"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySQLProvider MySQL数据库提供者
type MySQLProvider struct {
	db     *gorm.DB
	sqlDB  *sql.DB
	config Config
}

// NewMySQLProvider 创建MySQL提供者
func NewMySQLProvider(config Config) Provider {
	m := &MySQLProvider{
		config: config,
	}
	
	// 注册到注册表
	Register("mysql", func(cfg Config) Provider {
		return NewMySQLProvider(cfg)
	})
	
	return m
}

// Connect 连接数据库
func (m *MySQLProvider) Connect(ctx context.Context, config Config) error {
	dsn := m.buildDSN(config)
	
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.DefaultGormLogger(),
	})
	if err != nil {
		return fmt.Errorf("连接MySQL失败: %w", err)
	}
	
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取底层数据库连接失败: %w", err)
	}
	
	// 设置连接池参数
	if config.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	}
	if config.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	}
	
	m.db = db
	m.sqlDB = sqlDB
	
	// 测试连接
	if err := m.Ping(ctx); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}
	
	return nil
}

// Close 关闭连接
func (m *MySQLProvider) Close() error {
	if m.sqlDB != nil {
		return m.sqlDB.Close()
	}
	return nil
}

// Ping 测试连接
func (m *MySQLProvider) Ping(ctx context.Context) error {
	if m.sqlDB == nil {
		return ErrConnectionFailed
	}
	
	return m.sqlDB.PingContext(ctx)
}

// Find 查询多条记录
func (m *MySQLProvider) Find(ctx context.Context, dest any, query Query) error {
	db := m.applyQuery(ctx, query)
	return db.Find(dest).Error
}

// FindOne 查询单条记录
func (m *MySQLProvider) FindOne(ctx context.Context, dest any, query Query) error {
	db := m.applyQuery(ctx, query)
	return db.First(dest).Error
}

// Create 创建记录
func (m *MySQLProvider) Create(ctx context.Context, model any) error {
	return m.db.WithContext(ctx).Create(model).Error
}

// Update 更新记录
func (m *MySQLProvider) Update(ctx context.Context, model any, updates any) error {
	return m.db.WithContext(ctx).Model(model).Updates(updates).Error
}

// Delete 删除记录
func (m *MySQLProvider) Delete(ctx context.Context, models ...any) error {
	for _, model := range models {
		if err := m.db.WithContext(ctx).Delete(model).Error; err != nil {
			return err
		}
	}
	return nil
}

// Transaction 执行事务
func (m *MySQLProvider) Transaction(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		mysqlTx := &MySQLTx{db: tx}
		return fn(ctx, mysqlTx)
	})
}

// Exec 执行原生SQL
func (m *MySQLProvider) Exec(ctx context.Context, query string, args ...any) (Result, error) {
	result := m.db.WithContext(ctx).Exec(query, args...)
	if result.Error != nil {
		return Result{}, result.Error
	}
	
	return Result{
		RowsAffected: result.RowsAffected,
		LastInsertID: 0, // gorm.Exec 不返回 LastInsertId
	}, nil
}

// Query 执行查询SQL
func (m *MySQLProvider) Query(ctx context.Context, query string, args ...any) (Rows, error) {
	rows, err := m.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	
	return &MySQLRows{rows: rows}, nil
}

// Migrator 返回迁移器
func (m *MySQLProvider) Migrator() Migrator {
	return &MySQLMigrator{db: m.db}
}

// Builder 返回查询构建器
func (m *MySQLProvider) Builder() QueryBuilder {
	return &MySQLQueryBuilder{db: m.db}
}

// Dialect 返回方言
func (m *MySQLProvider) Dialect() Dialect {
	return &MySQLDialect{}
}

// buildDSN 构建MySQL连接字符串
func (m *MySQLProvider) buildDSN(config Config) string {
	if config.Charset == "" {
		config.Charset = "utf8mb4"
	}
	if config.Collation == "" {
		config.Collation = "utf8mb4_general_ci"
	}
	
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&collation=%s&parseTime=True&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.Charset,
		config.Collation,
	)
	
	// 添加额外参数
	if len(config.Params) > 0 {
		for key, value := range config.Params {
			dsn += fmt.Sprintf("&%s=%v", key, value)
		}
	}
	
	return dsn
}

// applyQuery 应用查询条件
func (m *MySQLProvider) applyQuery(ctx context.Context, query Query) *gorm.DB {
	db := m.db.WithContext(ctx)
	
	// where条件
	if len(query.Where) > 0 {
		for _, condition := range query.Where {
			switch strings.ToLower(condition.Operator) {
			case "=":
				db = db.Where(fmt.Sprintf("%s = ?", condition.Column), condition.Value)
			case "!=":
				db = db.Where(fmt.Sprintf("%s != ?", condition.Column), condition.Value)
			case ">":
				db = db.Where(fmt.Sprintf("%s > ?", condition.Column), condition.Value)
			case "<":
				db = db.Where(fmt.Sprintf("%s < ?", condition.Column), condition.Value)
			case ">=":
				db = db.Where(fmt.Sprintf("%s >= ?", condition.Column), condition.Value)
			case "<=":
				db = db.Where(fmt.Sprintf("%s <= ?", condition.Column), condition.Value)
			case "like":
				db = db.Where(fmt.Sprintf("%s LIKE ?", condition.Column), condition.Value)
			case "in":
				db = db.Where(fmt.Sprintf("%s IN ?", condition.Column), condition.Value)
			case "not_in":
				db = db.Where(fmt.Sprintf("%s NOT IN ?", condition.Column), condition.Value)
			case "is_null":
				db = db.Where(fmt.Sprintf("%s IS NULL", condition.Column))
			case "is_not_null":
				db = db.Where(fmt.Sprintf("%s IS NOT NULL", condition.Column))
			default:
				db = db.Where(fmt.Sprintf("%s %s ?", condition.Column, condition.Operator), condition.Value)
			}
		}
	}
	
	// order by
	for _, order := range query.Order {
		db = db.Order(fmt.Sprintf("%s %s", order.Column, strings.ToUpper(order.Direction)))
	}
	
	// limit
	if query.Limit != nil {
		db = db.Limit(*query.Limit)
	}
	
	// offset
	if query.Offset != nil {
		db = db.Offset(*query.Offset)
	}
	
	// joins
	for _, join := range query.Joins {
		db = db.Joins(join)
	}
	
	// select
	if len(query.Select) > 0 {
		db = db.Select(strings.Join(query.Select, ", "))
	}
	
	// group by
	if len(query.GroupBy) > 0 {
		db = db.Group(strings.Join(query.GroupBy, ", "))
	}
	
	// having
	if len(query.Having) > 0 {
		for _, condition := range query.Having {
			db = db.Having(fmt.Sprintf("%s %s ?", condition.Column, condition.Operator), condition.Value)
		}
	}
	
	return db
}

// MySQLTx MySQL事务实现
type MySQLTx struct {
	db *gorm.DB
}

// Find 查询多条记录
func (t *MySQLTx) Find(ctx context.Context, dest any, query Query) error {
	provider := &MySQLProvider{db: t.db}
	return provider.Find(ctx, dest, query)
}

// FindOne 查询单条记录
func (t *MySQLTx) FindOne(ctx context.Context, dest any, query Query) error {
	provider := &MySQLProvider{db: t.db}
	return provider.FindOne(ctx, dest, query)
}

// Create 创建记录
func (t *MySQLTx) Create(ctx context.Context, model any) error {
	provider := &MySQLProvider{db: t.db}
	return provider.Create(ctx, model)
}

// Update 更新记录
func (t *MySQLTx) Update(ctx context.Context, model any, updates any) error {
	provider := &MySQLProvider{db: t.db}
	return provider.Update(ctx, model, updates)
}

// Delete 删除记录
func (t *MySQLTx) Delete(ctx context.Context, models ...any) error {
	provider := &MySQLProvider{db: t.db}
	return provider.Delete(ctx, models...)
}

// Commit 提交事务
func (t *MySQLTx) Commit() error {
	return t.db.Commit().Error
}

// Rollback 回滚事务
func (t *MySQLTx) Rollback() error {
	return t.db.Rollback().Error
}

// Exec 执行原生SQL
func (t *MySQLTx) Exec(ctx context.Context, query string, args ...any) (Result, error) {
	provider := &MySQLProvider{db: t.db}
	return provider.Exec(ctx, query, args...)
}

// Query 执行查询SQL
func (t *MySQLTx) Query(ctx context.Context, query string, args ...any) (Rows, error) {
	provider := &MySQLProvider{db: t.db}
	return provider.Query(ctx, query, args...)
}

// MySQLRows MySQL结果集实现
type MySQLRows struct {
	rows *sql.Rows
}

func (r *MySQLRows) Close() error {
	return r.rows.Close()
}

func (r *MySQLRows) Next() bool {
	return r.rows.Next()
}

func (r *MySQLRows) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

func (r *MySQLRows) Columns() ([]string, error) {
	return r.rows.Columns()
}