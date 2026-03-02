// Package main 入口，仅负责组装与启动（规范 5.1、第九章）
//
// @title           Go Backend Framework API
// @version         1.0.0
// @description     通用企业级 Go 后端框架 - 符合开发规范 v2.0
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-backend-framework/api"
	_ "go-backend-framework/docs" // 注册 Swagger 文档，供 /swagger/* 使用
	"go-backend-framework/config"
	"go-backend-framework/pkg/framework"
	"go-backend-framework/pkg/logger"

	"go.uber.org/zap"
)

var (
	configPath = flag.String("config", "config/config.dev.yaml", "配置文件路径")
	version    = "1.0.0"
)

func main() {
	flag.Parse()

	// 1. 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "配置校验失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	if err := logger.Init(cfg.Log); err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	logger.Global().Info("应用启动", zap.String("version", version))

	// 3. 初始化 DB
	db, err := framework.NewMySQL(framework.MySQLConfig{
		Host:         cfg.MySQL.Host,
		Port:         cfg.MySQL.Port,
		User:         cfg.MySQL.User,
		Password:     cfg.MySQL.Password,
		Database:     cfg.MySQL.Database,
		Charset:      cfg.MySQL.Charset,
		Collation:    cfg.MySQL.Collation,
		MaxOpenConns: cfg.MySQL.MaxOpenConns,
		MaxIdleConns: cfg.MySQL.MaxIdleConns,
	})
	if err != nil {
		logger.Global().Fatal("连接数据库失败", zap.Error(err))
	}

	// 4. 初始化 Redis
	rdb, err := framework.NewRedis(framework.RedisConfig{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		logger.Global().Fatal("连接 Redis 失败", zap.Error(err))
	}

	// 5. 组装路由
	engine, err := api.Setup(cfg, db, rdb)
	if err != nil {
		logger.Global().Fatal("组装路由失败", zap.Error(err))
	}

	// 6. 启动 HTTP 服务
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	go func() {
		logger.Global().Info("HTTP 服务启动", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Global().Error("HTTP 服务异常", zap.Error(err))
		}
	}()

	// 7. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Global().Info("正在关闭服务...")
	if err := srv.Shutdown(ctx); err != nil {
		logger.Global().Error("服务关闭异常", zap.Error(err))
	}

	// 关闭 Redis
	_ = rdb.Close()

	logger.Global().Info("服务已关闭")
}
