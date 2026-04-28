// Package database SQLx 数据库提供者（可选，高性能实现）
package database

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func init() {
	Register("mysql-sqlx", func(cfg Config) Provider {
		return NewSQLxProvider(cfg)
	})
}

// SQLxProvider 基于 SQLx 的数据库提供者
type SQLxProvider struct {
	db     *sqlx.DB
	config Config
}

// NewSQLxProvider 创建 SQLx 提供者
func NewSQLxProvider(config Config) Provider {
	return &SQLxProvider{config: config}
}

// Connect 连接
func (s *SQLxProvider) Connect(ctx context.Context, config Config) error {
	dsn := buildDSN(config)
	db, err := sqlx.ConnectContext(ctx, "mysql", dsn)
	if err != nil {
		return fmt.Errorf("sqlx 连接失败: %w", err)
	}
	if config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(config.MaxIdleConns)
	}
	s.db = db
	return s.Ping(ctx)
}

func buildDSN(c Config) string {
	if c.Charset == "" {
		c.Charset = "utf8mb4"
	}
	if c.Collation == "" {
		c.Collation = "utf8mb4_general_ci"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&collation=%s&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset, c.Collation)
}

// Close 关闭
func (s *SQLxProvider) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Ping 测试连接
func (s *SQLxProvider) Ping(ctx context.Context) error {
	if s.db == nil {
		return ErrConnectionFailed
	}
	return s.db.PingContext(ctx)
}

// Find 查询多条
func (s *SQLxProvider) Find(ctx context.Context, dest any, query Query) error {
	table := getTableFromDest(dest, query.Table)
	if table == "" {
		return fmt.Errorf("query.Table 或 model 需提供表名")
	}
	sqlStr, args := BuildSelectSQL(table, query)
	return s.db.SelectContext(ctx, dest, sqlStr, args...)
}

// FindOne 查询单条
func (s *SQLxProvider) FindOne(ctx context.Context, dest any, query Query) error {
	table := getTableFromDest(dest, query.Table)
	if table == "" {
		return fmt.Errorf("query.Table 或 model 需提供表名")
	}
	// 强制 LIMIT 1
	q := query
	one := 1
	q.Limit = &one
	sqlStr, args := BuildSelectSQL(table, q)
	return s.db.GetContext(ctx, dest, sqlStr, args...)
}

func getTableFromDest(dest any, hint string) string {
	if hint != "" {
		return hint
	}
	// 尝试从 model 的 TableName() 获取
	t := reflect.TypeOf(dest)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}
	m, ok := reflect.New(t).Interface().(interface{ TableName() string })
	if ok {
		return m.TableName()
	}
	return ""
}

// Create 创建（需 model 有 db 标签）
func (s *SQLxProvider) Create(ctx context.Context, model any) error {
	table := getTableFromDest(model, "")
	if table == "" {
		return fmt.Errorf("model 需实现 TableName()")
	}
	query, args, err := sqlx.Named("INSERT INTO "+table+" ("+insertColumns(model)+") VALUES ("+insertPlaceholders(model)+")", structToMap(model))
	if err != nil {
		return err
	}
	query = s.db.Rebind(query)
	_, err = s.db.ExecContext(ctx, query, args...)
	return err
}

func structToMap(v any) map[string]interface{} {
	// 简化：使用 sqlx 的 Map 能力
	m := make(map[string]interface{})
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		f := typ.Field(i)
		if f.PkgPath != "" {
			continue
		}
		dbTag := f.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}
		col := strings.Split(dbTag, ",")[0]
		m[col] = val.Field(i).Interface()
	}
	return m
}

func insertColumns(v any) string {
	var cols []string
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		f := typ.Field(i)
		dbTag := f.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}
		col := strings.Split(dbTag, ",")[0]
		if col == "id" && isAutoIncrement(val.Field(i)) {
			continue
		}
		cols = append(cols, col)
	}
	return strings.Join(cols, ", ")
}

func insertPlaceholders(v any) string {
	var ph []string
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		f := typ.Field(i)
		dbTag := f.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}
		col := strings.Split(dbTag, ",")[0]
		if col == "id" && isAutoIncrement(val.Field(i)) {
			continue
		}
		ph = append(ph, ":"+col)
	}
	return strings.Join(ph, ", ")
}

func isAutoIncrement(v reflect.Value) bool {
	if v.Kind() == reflect.Int64 || v.Kind() == reflect.Int {
		return v.Int() == 0
	}
	return false
}

// Update 更新
func (s *SQLxProvider) Update(ctx context.Context, model any, updates any) error {
	table := getTableFromDest(model, "")
	if table == "" {
		return fmt.Errorf("model 需实现 TableName()")
	}
	m, ok := updates.(map[string]interface{})
	if !ok {
		m = structToMap(updates)
	}
	var setParts []string
	var args []any
	for k, v := range m {
		if k == "id" {
			continue
		}
		setParts = append(setParts, k+" = ?")
		args = append(args, v)
	}
	id := getIDFromModel(model)
	if id == nil {
		return fmt.Errorf("model 需有 id 字段")
	}
	args = append(args, id)
	sqlStr := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", table, strings.Join(setParts, ", "))
	_, err := s.db.ExecContext(ctx, sqlStr, args...)
	return err
}

func getIDFromModel(v any) interface{} {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	f := val.FieldByName("ID")
	if !f.IsValid() {
		return nil
	}
	return f.Interface()
}

// Delete 删除
func (s *SQLxProvider) Delete(ctx context.Context, models ...any) error {
	for _, m := range models {
		table := getTableFromDest(m, "")
		if table == "" {
			return fmt.Errorf("model 需实现 TableName()")
		}
		id := getIDFromModel(m)
		if id == nil {
			return fmt.Errorf("model 需有 id 字段")
		}
		_, err := s.db.ExecContext(ctx, "DELETE FROM "+table+" WHERE id = ?", id)
		if err != nil {
			return err
		}
	}
	return nil
}

// Transaction 事务
func (s *SQLxProvider) Transaction(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	sqlxTx := &SQLxTx{tx: tx}
	if err := fn(ctx, sqlxTx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// Exec 原生执行
func (s *SQLxProvider) Exec(ctx context.Context, query string, args ...any) (Result, error) {
	r, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return Result{}, err
	}
	rows, _ := r.RowsAffected()
	lastID, _ := r.LastInsertId()
	return Result{RowsAffected: rows, LastInsertID: lastID}, nil
}

// Query 原生查询
func (s *SQLxProvider) Query(ctx context.Context, query string, args ...any) (Rows, error) {
	rows, err := s.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &sqlRowsWrapper{rows: rows}, nil
}

// Migrator 迁移器（SQLx 无内置，返回 noop）
func (s *SQLxProvider) Migrator() Migrator {
	return &sqlxMigrator{db: s.db}
}

// Builder 查询构建器
func (s *SQLxProvider) Builder() QueryBuilder {
	return &MySQLQueryBuilder{query: Query{}}
}

// Dialect 方言
func (s *SQLxProvider) Dialect() Dialect {
	return &MySQLDialect{}
}

type SQLxTx struct {
	tx *sqlx.Tx
}

func (t *SQLxTx) Find(ctx context.Context, dest any, query Query) error {
	table := getTableFromDest(dest, query.Table)
	if table == "" {
		return fmt.Errorf("query.Table 或 model 需提供表名")
	}
	sqlStr, args := BuildSelectSQL(table, query)
	return t.tx.SelectContext(ctx, dest, sqlStr, args...)
}

func (t *SQLxTx) FindOne(ctx context.Context, dest any, query Query) error {
	table := getTableFromDest(dest, query.Table)
	if table == "" {
		return fmt.Errorf("query.Table 或 model 需提供表名")
	}
	q := query
	one := 1
	q.Limit = &one
	sqlStr, args := BuildSelectSQL(table, q)
	return t.tx.GetContext(ctx, dest, sqlStr, args...)
}

func (t *SQLxTx) Create(ctx context.Context, model any) error {
	table := getTableFromDest(model, "")
	if table == "" {
		return fmt.Errorf("model 需实现 TableName()")
	}
	query, args, err := sqlx.Named("INSERT INTO "+table+" ("+insertColumns(model)+") VALUES ("+insertPlaceholders(model)+")", structToMap(model))
	if err != nil {
		return err
	}
	query = t.tx.Rebind(query)
	_, err = t.tx.ExecContext(ctx, query, args...)
	return err
}

func (t *SQLxTx) Update(ctx context.Context, model any, updates any) error {
	table := getTableFromDest(model, "")
	if table == "" {
		return fmt.Errorf("model 需实现 TableName()")
	}
	m, ok := updates.(map[string]interface{})
	if !ok {
		m = structToMap(updates)
	}
	var setParts []string
	var args []any
	for k, v := range m {
		if k == "id" {
			continue
		}
		setParts = append(setParts, k+" = ?")
		args = append(args, v)
	}
	id := getIDFromModel(model)
	if id == nil {
		return fmt.Errorf("model 需有 id 字段")
	}
	args = append(args, id)
	sqlStr := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", table, strings.Join(setParts, ", "))
	_, err := t.tx.ExecContext(ctx, sqlStr, args...)
	return err
}

func (t *SQLxTx) Delete(ctx context.Context, models ...any) error {
	for _, m := range models {
		table := getTableFromDest(m, "")
		if table == "" {
			return fmt.Errorf("model 需实现 TableName()")
		}
		id := getIDFromModel(m)
		if id == nil {
			return fmt.Errorf("model 需有 id 字段")
		}
		if _, err := t.tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE id = ?", id); err != nil {
			return err
		}
	}
	return nil
}

func (t *SQLxTx) Commit() error {
	return t.tx.Commit()
}

func (t *SQLxTx) Rollback() error {
	return t.tx.Rollback()
}

func (t *SQLxTx) Exec(ctx context.Context, query string, args ...any) (Result, error) {
	r, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return Result{}, err
	}
	rows, _ := r.RowsAffected()
	lastID, _ := r.LastInsertId()
	return Result{RowsAffected: rows, LastInsertID: lastID}, nil
}

func (t *SQLxTx) Query(ctx context.Context, query string, args ...any) (Rows, error) {
	rows, err := t.tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &sqlRowsWrapper{rows: rows}, nil
}

type sqlRowsWrapper struct {
	rows *sqlx.Rows
}

func (r *sqlRowsWrapper) Close() error {
	return r.rows.Close()
}

func (r *sqlRowsWrapper) Next() bool {
	return r.rows.Next()
}

func (r *sqlRowsWrapper) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

func (r *sqlRowsWrapper) Columns() ([]string, error) {
	return r.rows.Columns()
}

type sqlxMigrator struct {
	db *sqlx.DB
}

func (m *sqlxMigrator) CreateTable(ctx context.Context, model any) error {
	return fmt.Errorf("sqlx migrator: 请使用 Gorm 或独立迁移工具")
}

func (m *sqlxMigrator) DropTable(ctx context.Context, model any) error {
	return fmt.Errorf("sqlx migrator: 请使用 Gorm 或独立迁移工具")
}

func (m *sqlxMigrator) HasTable(ctx context.Context, tableName string) (bool, error) {
	var n int
	err := m.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", tableName).Scan(&n)
	return n > 0, err
}

func (m *sqlxMigrator) AddColumn(ctx context.Context, tableName, column string, fieldType interface{}) error {
	_, err := m.db.ExecContext(ctx, "ALTER TABLE "+tableName+" ADD COLUMN "+column+" "+fmt.Sprint(fieldType))
	return err
}

func (m *sqlxMigrator) DropColumn(ctx context.Context, tableName, column string) error {
	_, err := m.db.ExecContext(ctx, "ALTER TABLE "+tableName+" DROP COLUMN "+column)
	return err
}

func (m *sqlxMigrator) AutoMigrate(ctx context.Context, models ...any) error {
	return fmt.Errorf("sqlx migrator: 请使用 Gorm 或独立迁移工具")
}
