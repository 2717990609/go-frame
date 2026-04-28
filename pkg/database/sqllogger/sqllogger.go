// Package sqllogger 数据库 SQL 日志，实现 GORM logger.Interface
//
// 支持 trace_id 透传（从 context 提取，由 logger.C(ctx) 注入），
// 可被 MySQL、PostgreSQL 等 GORM 数据库插件复用。
//
// 使用方式：在创建 gorm.DB 时传入
//   gormConfig := &gorm.Config{}
//   if enableSQLLog { gormConfig.Logger = sqllogger.New(true) }
//   db, _ := gorm.Open(postgres.Open(dsn), gormConfig)
package sqllogger

import (
	"context"
	"fmt"
	"time"

	"go-backend-framework/pkg/logger"
	"go.uber.org/zap"
	gormlogger "gorm.io/gorm/logger"
)

// New 创建支持 trace_id 的 GORM SQL Logger
// enable 为 true 时打印 SQL；为 false 时使用 Silent 模式
func New(enable bool) gormlogger.Interface {
	if !enable {
		return gormlogger.Default.LogMode(gormlogger.Silent)
	}
	l := &sqlLogger{}
	return l.LogMode(gormlogger.Info)
}

// sqlLogger 实现 gorm/logger.Interface，从 context 提取 trace_id 并输出到 zap
type sqlLogger struct {
	level gormlogger.LogLevel
}

func (l *sqlLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	nl := *l
	nl.level = level
	return &nl
}

func (l *sqlLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	l.logInfo(ctx, msg, args...)
}

func (l *sqlLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.logWarn(ctx, msg, args...)
}

func (l *sqlLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	l.logError(ctx, msg, args...)
}

func (l *sqlLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sqlStr, rows := fc()
	zlog := logger.C(ctx)
	fields := []zap.Field{
		zap.Duration("duration", elapsed),
		zap.String("sql", sqlStr),
		zap.Int64("rows", rows),
	}
	if err != nil {
		zlog.Error("sql execute error", append(fields, zap.Error(err))...)
		return
	}
	if l.level >= gormlogger.Info {
		zlog.Info("sql execute", fields...)
	}
}

func (l *sqlLogger) logInfo(ctx context.Context, msg string, args ...interface{}) {
	if l.level <= gormlogger.Silent {
		return
	}
	logger.C(ctx).Info(msg, toZapFields(args...)...)
}

func (l *sqlLogger) logWarn(ctx context.Context, msg string, args ...interface{}) {
	if l.level <= gormlogger.Silent {
		return
	}
	logger.C(ctx).Warn(msg, toZapFields(args...)...)
}

func (l *sqlLogger) logError(ctx context.Context, msg string, args ...interface{}) {
	if l.level <= gormlogger.Silent {
		return
	}
	logger.C(ctx).Error(msg, toZapFields(args...)...)
}

func toZapFields(args ...interface{}) []zap.Field {
	fields := make([]zap.Field, 0, (len(args)+1)/2)
	for i := 0; i+1 < len(args); i += 2 {
		key := fmt.Sprintf("%v", args[i])
		fields = append(fields, zap.Any(key, args[i+1]))
	}
	return fields
}
