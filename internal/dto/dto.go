// Package dto Handler 层 Request/Response 结构体（规范 5.2）
package dto

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
}

// ReadyResponse 就绪检查响应，含依赖状态
type ReadyResponse struct {
	Ready   bool              `json:"ready"`
	MySQL   bool              `json:"mysql"`
	Redis   bool              `json:"redis"`
	Details map[string]string `json:"details,omitempty"`
}
