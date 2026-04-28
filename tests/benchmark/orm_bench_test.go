// Package benchmark ORM 性能基准测试（Phase 2 技术栈升级）
// 使用 SQLite 内存库，无需外部 MySQL
package benchmark

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// benchUser 基准测试用模型
type benchUser struct {
	ID    int64  `gorm:"column:id;primaryKey" db:"id"`
	Email string `gorm:"column:email" db:"email"`
}

func (benchUser) TableName() string { return "bench_users" }

func setupGormDB(b *testing.B) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		b.Skipf("sqlite not available (need CGO): %v", err)
	}
	if err := db.AutoMigrate(&benchUser{}); err != nil {
		b.Fatal(err)
	}
	return db
}

func setupSQLxDB(b *testing.B) *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		b.Skipf("sqlite not available: %v", err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bench_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT
		)
	`)
	if err != nil {
		b.Fatal(err)
	}
	return db
}

// BenchmarkGorm_Insert  Gorm 插入
func BenchmarkGorm_Insert(b *testing.B) {
	db := setupGormDB(b)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		u := &benchUser{Email: "test@example.com"}
		_ = db.WithContext(ctx).Create(u).Error
	}
}

// BenchmarkSQLx_Insert SQLx 插入
func BenchmarkSQLx_Insert(b *testing.B) {
	db := setupSQLxDB(b)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = db.ExecContext(ctx, "INSERT INTO bench_users (email) VALUES (?)", "test@example.com")
	}
}

// BenchmarkGorm_Find Gorm 查询
func BenchmarkGorm_Find(b *testing.B) {
	db := setupGormDB(b)
	ctx := context.Background()
	_ = db.WithContext(ctx).Create(&benchUser{Email: "find@example.com"}).Error
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var u benchUser
		_ = db.WithContext(ctx).Where("email = ?", "find@example.com").First(&u).Error
	}
}

// BenchmarkSQLx_Get SQLx Get 查询
func BenchmarkSQLx_Get(b *testing.B) {
	db := setupSQLxDB(b)
	ctx := context.Background()
	_, _ = db.ExecContext(ctx, "INSERT INTO bench_users (email) VALUES (?)", "find@example.com")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var u benchUser
		_ = db.GetContext(ctx, &u, "SELECT id, email FROM bench_users WHERE email = ?", "find@example.com")
	}
}

// BenchmarkDatabaseSQL_Query 原生 database/sql 查询
func BenchmarkDatabaseSQL_Query(b *testing.B) {
	sqldb, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Skipf("sqlite not available: %v", err)
	}
	defer sqldb.Close()
	_, _ = sqldb.Exec("CREATE TABLE IF NOT EXISTS bench_users (id INTEGER PRIMARY KEY, email TEXT)")
	_, _ = sqldb.Exec("INSERT INTO bench_users (id, email) VALUES (1, ?)", "raw@example.com")
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var id int64
		var email string
		_ = sqldb.QueryRowContext(ctx, "SELECT id, email FROM bench_users WHERE id = 1").Scan(&id, &email)
	}
}
