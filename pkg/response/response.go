// Package response 统一响应封装，符合《星火现梦》后端开发规范 2.2
package response

import (
	"time"
)

// Response 响应包装器，所有接口强制使用
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
	Ts   int64       `json:"ts"`
}

// Success 成功响应，data 禁止为 nil
func Success(data interface{}) Response {
	if data == nil {
		data = map[string]interface{}{}
	}
	return Response{
		Code: 200,
		Msg:  "success",
		Data: data,
		Ts:   time.Now().Unix(),
	}
}

// Error 错误响应，禁止返回 nil data
func Error(code int, msg string) Response {
	return Response{
		Code: code,
		Msg:  msg,
		Data: map[string]interface{}{},
		Ts:   time.Now().Unix(),
	}
}

// 错误码常量（规范 2.3）
const (
	CodeSuccess           = 200
	CodeParamError        = 4001 // 客户端参数错误
	CodeBusinessError     = 4101 // 业务逻辑错误
	CodePermissionError   = 4201 // 权限/状态错误
	CodeServerError       = 5001 // 服务端错误
	CodeThirdPartyError   = 6001 // 第三方服务错误
)
