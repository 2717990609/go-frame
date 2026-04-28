// Package logger 日志封装，支持 trace_id 全链路透传（规范 3.1）
package logger

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const traceIDKey = "trace_id"

type ctxKey struct{}

var defaultLogger *zap.Logger
var traceIDKeyType ctxKey

// Config 日志配置，对应主配置文件中的 log 段（如 config/config.dev.yaml）
type Config struct {
	Output          []string `yaml:"output"`
	FilePath        string   `yaml:"file_path"`
	MaxSize         int      `yaml:"max_size"` // MB
	MaxBackups      int      `yaml:"max_backups"`
	MaxAge          int      `yaml:"max_age"` // 天
	Compress        bool     `yaml:"compress"`
	RotationTime    string   `yaml:"rotation_time"`
	EnableSQLLog    bool     `yaml:"enable_sql_log"`
	SensitiveFields []string `yaml:"sensitive_fields"`
}

// Init 初始化日志，需在应用启动时调用
func Init(cfg Config) error {
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:  "msg",
		LevelKey:    "level",
		TimeKey:     "timestamp",
		EncodeLevel: zapcore.LowercaseLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05.000 -0700"))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		CallerKey:      "caller",
		EncodeCaller:   zapcore.ShortCallerEncoder, // 输出文件名和行号，如 pkg/xxx/file.go:123
	}
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	level := zapcore.DebugLevel

	var writeSyncer zapcore.WriteSyncer
	if len(cfg.FilePath) > 0 {
		writeSyncer = zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		})
	} else {
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	core := zapcore.NewCore(encoder, writeSyncer, level)
	defaultLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return nil
}

// SetLogger 设置全局 logger（用于测试或自定义配置）
func SetLogger(l *zap.Logger) {
	defaultLogger = l
}

// C 从 context 获取带 trace_id 的 logger，业务代码统一使用此方法
func C(ctx context.Context) *zap.Logger {
	if ctx == nil || defaultLogger == nil {
		return defaultLogger
	}
	traceID, ok := ctx.Value(traceIDKeyType).(string)
	if !ok || traceID == "" {
		return defaultLogger
	}
	return defaultLogger.With(zap.String(traceIDKey, traceID))
}

// WithTraceID 将 trace_id 注入 context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKeyType, traceID)
}

// Global 返回全局 logger（无 trace_id 场景）
func Global() *zap.Logger {
	if defaultLogger == nil {
		defaultLogger, _ = zap.NewProduction()
	}
	return defaultLogger
}
