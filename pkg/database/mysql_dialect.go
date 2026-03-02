// Package database MySQL方言实现
package database

import (
	"fmt"
	"strings"
	"time"
)

// MySQLDialect MySQL方言实现
type MySQLDialect struct{}

// Quote 标记标识符
func (d *MySQLDialect) Quote(identifier string) string {
	return fmt.Sprintf("`%s`", identifier)
}

// Now 获取当前时间函数
func (d *MySQLDialect) Now() string {
	return "NOW()"
}

// ColumnType 获取列类型
func (d *MySQLDialect) ColumnType(fieldType string) string {
	typeMap := map[string]string{
		"int":        "INT",
		"int64":      "BIGINT",
		"int32":      "INT",
		"int16":      "SMALLINT",
		"int8":       "TINYINT",
		"uint":       "INT UNSIGNED",
		"uint64":     "BIGINT UNSIGNED",
		"uint32":     "INT UNSIGNED",
		"uint16":     "SMALLINT UNSIGNED",
		"uint8":      "TINYINT UNSIGNED",
		"float32":    "FLOAT",
		"float64":    "DOUBLE",
		"string":     "VARCHAR(255)",
		"text":       "TEXT",
		"bool":       "TINYINT(1)",
		"time":       "DATETIME",
		"time.Time":  "DATETIME",
		"json":       "JSON",
		"blob":       "BLOB",
	}
	
	if sqlType, exists := typeMap[fieldType]; exists {
		return sqlType
	}
	
	// 默认返回VARCHAR
	return "VARCHAR(255)"
}

// SupportsLastInsertID 是否支持LastInsertID
func (d *MySQLDialect) SupportsLastInsertID() bool {
	return true
}

// SupportsReturningClause 是否支持RETURN子句
func (d *MySQLDialect) SupportsReturningClause() bool {
	// MySQL 8.0.16+ 支持RETURNING，但为了兼容性统一返回false
	return false
}

// MySQLMigrator MySQL迁移器实现
type MySQLMigrator struct {
	db *gorm.DB
}

// CreateTable 创建表
func (m *MySQLMigrator) CreateTable(ctx context.Context, model any) error {
	return m.db.WithContext(ctx).AutoMigrate(model)
}

// DropTable 删除表
func (m *MySQLMigrator) DropTable(ctx context.Context, model any) error {
	return m.db.WithContext(ctx).Migrator().DropTable(model).Error
}

// HasTable 检查表是否存在
func (m *MySQLMigrator) HasTable(ctx context.Context, tableName string) (bool, error) {
	var count int64
	err := m.db.WithContext(ctx).Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", tableName).Scan(&count).Error
	return count > 0, err
}

// AddColumn 添加列
func (m *MySQLMigrator) AddColumn(ctx context.Context, tableName, column string, fieldType interface{}) error {
	// 解析字段类型
	var sqlType string
	switch v := fieldType.(type) {
	case string:
		sqlType = v
	default:
		dialect := &MySQLDialect{}
		sqlType = dialect.ColumnType(fmt.Sprintf("%T", fieldType))
	}
	
	// 构建ALTER TABLE语句
	sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, column, sqlType)
	return m.db.WithContext(ctx).Exec(sql).Error
}

// DropColumn 删除列
func (m *MySQLMigrator) DropColumn(ctx context.Context, tableName, column string) error {
	sql := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tableName, column)
	return m.db.WithContext(ctx).Exec(sql).Error
}

// AutoMigrate 自动迁移
func (m *MySQLMigrator) AutoMigrate(ctx context.Context, models ...any) error {
	return m.db.WithContext(ctx).AutoMigrate(models...)
}

// MySQLQueryBuilder MySQL查询构建器实现
type MySQLQueryBuilder struct {
	db     *gorm.DB
	query  Query
	ctx    context.Context
}

// Select 选择字段
func (b *MySQLQueryBuilder) Select(columns ...string) QueryBuilder {
	b.query.Select = columns
	return b
}

// From 设置表名
func (b *MySQLQueryBuilder) From(table string) QueryBuilder {
	// MySQLBuilder中表名通过gorm的Model或Table方法设置
	// 这里简化处理
	return b
}

// Where 设置WHERE条件
func (b *MySQLQueryBuilder) Where(conditions ...Condition) QueryBuilder {
	b.query.Where = append(b.query.Where, conditions...)
	return b
}

// OrderBy 设置排序
func (b *MySQLQueryBuilder) OrderBy(column, direction string) QueryBuilder {
	b.query.Order = append(b.query.Order, Order{Column: column, Direction: direction})
	return b
}

// Limit 设置限制
func (b *MySQLQueryBuilder) Limit(limit int) QueryBuilder {
	b.query.Limit = &limit
	return b
}

// Offset 设置偏移
func (b *MySQLQueryBuilder) Offset(offset int) QueryBuilder {
	b.query.Offset = &offset
	return b
}

// Join 设置连接
func (b *MySQLQueryBuilder) Join(table string, on string) QueryBuilder {
	joinClause := fmt.Sprintf("JOIN %s ON %s", table, on)
	b.query.Joins = append(b.query.Joins, joinClause)
	return b
}

// GroupBy 设置分组
func (b *MySQLQueryBuilder) GroupBy(columns ...string) QueryBuilder {
	b.query.GroupBy = columns
	return b
}

// Having 设置HAVING条件
func (b *MySQLQueryBuilder) Having(conditions ...Condition) QueryBuilder {
	b.query.Having = append(b.query.Having, conditions...)
	return b
}

// Build 构建SQL语句
func (b *MySQLQueryBuilder) Build() (string, []any) {
	var args []any
	var parts []string
	
	// SELECT部分
	if len(b.query.Select) > 0 {
		parts = append(parts, "SELECT "+strings.Join(b.query.Select, ", "))
	} else {
		parts = append(parts, "SELECT *")
	}
	
	// FROM部分（这里简化，实际应该提供Model或Table）
	parts = append(parts, "FROM unknown_table") // 需要调用者提供表名
	
	// WHERE部分
	if len(b.query.Where) > 0 {
		whereParts := make([]string, len(b.query.Where))
		for i, condition := range b.query.Where {
			switch strings.ToLower(condition.Operator) {
			case "=":
				whereParts[i] = fmt.Sprintf("%s = ?", condition.Column)
			case "!=":
				whereParts[i] = fmt.Sprintf("%s != ?", condition.Column)
			case "like":
				whereParts[i] = fmt.Sprintf("%s LIKE ?", condition.Column)
			case "in":
				whereParts[i] = fmt.Sprintf("%s IN ?", condition.Column)
			case "is_null":
				whereParts[i] = fmt.Sprintf("%s IS NULL", condition.Column)
			case "is_not_null":
				whereParts[i] = fmt.Sprintf("%s IS NOT NULL", condition.Column)
			default:
				whereParts[i] = fmt.Sprintf("%s %s ?", condition.Column, condition.Operator)
			}
			
			if condition.Value != nil {
				args = append(args, condition.Value)
			}
		}
		parts = append(parts, "WHERE "+strings.Join(whereParts, " AND "))
	}
	
	// ORDER BY部分
	if len(b.query.Order) > 0 {
		orderParts := make([]string, len(b.query.Order))
		for i, order := range b.query.Order {
			orderParts[i] = fmt.Sprintf("%s %s", order.Column, strings.ToUpper(order.Direction))
		}
		parts = append(parts, "ORDER BY "+strings.Join(orderParts, ", "))
	}
	
	// LIMIT部分
	if b.query.Limit != nil {
		parts = append(parts, fmt.Sprintf("LIMIT %d", *b.query.Limit))
	}
	
	// OFFSET部分
	if b.query.Offset != nil {
		parts = append(parts, fmt.Sprintf("OFFSET %d", *b.query.Offset))
	}
	
	return strings.Join(parts, " "), args
}