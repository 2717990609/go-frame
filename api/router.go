// Package api 路由组装
package api

import (
	"time"

	"go-backend-framework/config"
	"go-backend-framework/internal/handler"
	"go-backend-framework/internal/service"
	"go-backend-framework/pkg/middleware"
	"go-backend-framework/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// Setup 组装路由与中间件（规范第七章中间件链顺序）
func Setup(cfg *config.Config, db *gorm.DB, rdb *redis.Client) (*gin.Engine, error) {
	r := gin.New()

	// 1. Recovery
	r.Use(middleware.Recovery())
	// 2. RequestID
	r.Use(middleware.RequestID())
	// 3. Logger
	r.Use(middleware.Logger())
	// 4. Timeout
	timeout := time.Duration(cfg.Server.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	r.Use(middleware.Timeout(timeout))
	// 5. CORS
	r.Use(middleware.CORS(cfg.CORS.OriginsSlice()))

	// 全局限流（示例：100 req/min）
	limiter := middleware.NewRateLimiter(100, time.Minute)
	r.Use(middleware.RateLimit(limiter))

	// 健康检查（无认证）
	healthSvc := service.NewHealthService(db, rdb)
	healthH := handler.NewHealthHandler(healthSvc)
	r.GET("/health", healthH.Health)
	r.GET("/ready", healthH.Ready)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, response.Success(map[string]string{"pong": "ok"}))
	})

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		signed := v1.Group("")
		signed.Use(middleware.Signature(cfg.Signature, rdb))
		signed.POST("/echo", func(c *gin.Context) {
			var body map[string]interface{}
			_ = c.ShouldBindJSON(&body)
			c.JSON(200, response.Success(body))
		})
	}

	return r, nil
}
