// Package handler HTTP 层，参数校验、调用 Service（规范 5.1）
package handler

import (
	"net/http"

	"go-backend-framework/internal/dto"
	"go-backend-framework/internal/service"
	"go-backend-framework/pkg/response"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查
type HealthHandler struct {
	svc *service.HealthService
}

// 用于 swag 解析 @Success 中的 dto 类型
var _ = (*dto.HealthResponse)(nil)

// NewHealthHandler 创建
func NewHealthHandler(svc *service.HealthService) *HealthHandler {
	return &HealthHandler{svc: svc}
}

// Health godoc
// @Summary      存活探测
// @Description  用于 K8s 存活探针，不检查依赖
// @Tags         健康检查
// @Produce      json
// @Success      200  {object}  response.Response{data=dto.HealthResponse}
// @Router       /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	resp := h.svc.Health(c.Request.Context())
	c.JSON(http.StatusOK, response.Success(resp))
}

// Ready godoc
// @Summary      就绪探测
// @Description  用于负载均衡摘除，检查 MySQL、Redis 连通性
// @Tags         健康检查
// @Produce      json
// @Success      200  {object}  response.Response{data=dto.ReadyResponse}
// @Router       /ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	resp := h.svc.Ready(c.Request.Context())
	if !resp.Ready {
		c.JSON(http.StatusServiceUnavailable, response.Success(resp))
		return
	}
	c.JSON(http.StatusOK, response.Success(resp))
}
