// Package middleware 请求验签中间件
package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-backend-framework/config"
	"go-backend-framework/pkg/logger"
	"go-backend-framework/pkg/response"
	"go-backend-framework/pkg/signature"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	headerSignature = "X-Signature"
	headerTimestamp = "X-Timestamp"
	headerNonce     = "X-Nonce"
	headerEncrypted = "X-Encrypted"
)

// Signature 验签中间件，支持 Query/Form/JSON 参数及可选 Body 解密
func Signature(cfg config.SignatureConfig, rdb *redis.Client) gin.HandlerFunc {
	tw := time.Duration(cfg.TimeWindow) * time.Second
	if tw <= 0 {
		tw = 5 * time.Minute
	}
	nt := time.Duration(cfg.NonceTTL) * time.Second
	if nt <= 0 {
		nt = 10 * time.Second
	}
	return func(c *gin.Context) {
		if !cfg.Enabled || cfg.Key == "" {
			c.Next()
			return
		}

		// 1. 收集参数
		params := collectParams(c, cfg, rdb)
		if params == nil {
			return
		}

		// 2. 必填 Header
		sign := c.GetHeader(headerSignature)
		timestamp := c.GetHeader(headerTimestamp)
		nonce := c.GetHeader(headerNonce)
		if sign == "" || timestamp == "" || nonce == "" {
			logger.C(c.Request.Context()).Warn("验签失败：缺少签名头",
				zap.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusBadRequest, response.Error(response.CodeParamError, "缺少签名参数，请传递 X-Signature、X-Timestamp、X-Nonce"))
			c.Abort()
			return
		}

		params["timestamp"] = timestamp
		params["nonce"] = nonce

		// 3. 时间戳窗口
		if err := signature.ValidateTimestamp(timestamp, tw); err != nil {
			logger.C(c.Request.Context()).Warn("验签失败：时间戳超时",
				zap.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusBadRequest, response.Error(response.CodeParamError, err.Error()))
			c.Abort()
			return
		}

		// 4. 防重放
		if rdb != nil {
			nonceKey := "sign:nonce:" + nonce
			ok, err := rdb.SetNX(c.Request.Context(), nonceKey, "1", nt).Result()
			if err != nil {
				logger.C(c.Request.Context()).Error("验签防重放 Redis 异常", zap.Error(err))
				c.JSON(http.StatusInternalServerError, response.Error(response.CodeServerError, "系统繁忙，请稍后重试"))
				c.Abort()
				return
			}
			if !ok {
				logger.C(c.Request.Context()).Warn("验签失败：重复请求", zap.String("nonce", nonce))
				c.JSON(http.StatusBadRequest, response.Error(response.CodeParamError, "请勿重复提交"))
				c.Abort()
				return
			}
		}

		// 5. 验签
		if !signature.Verify(params, sign, cfg.Key) {
			logger.C(c.Request.Context()).Warn("验签失败：签名不匹配",
				zap.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusBadRequest, response.Error(response.CodeParamError, "签名验证失败"))
			c.Abort()
			return
		}

		c.Next()
	}
}

func collectParams(c *gin.Context, cfg config.SignatureConfig, rdb *redis.Client) map[string]string {
	params := make(map[string]string)

	// Query
	for k, v := range c.Request.URL.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	// Form
	_ = c.Request.ParseForm()
	for k, v := range c.Request.PostForm {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	contentType := strings.ToLower(c.GetHeader("Content-Type"))
	isJSON := strings.Contains(contentType, "application/json")

	if isJSON {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.Error(response.CodeParamError, "请求体读取失败"))
			c.Abort()
			return nil
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// 若加密，先解密
		if strings.ToLower(c.GetHeader(headerEncrypted)) == "true" && len(cfg.Key) > 0 {
			plain, err := signature.DecryptBody(bodyBytes, true, []byte(cfg.Key))
			if err != nil {
				logger.C(c.Request.Context()).Warn("解密失败", zap.Error(err))
				c.JSON(http.StatusBadRequest, response.Error(response.CodeParamError, "参数解密失败"))
				c.Abort()
				return nil
			}
			bodyBytes = plain
			c.Request.Body = io.NopCloser(bytes.NewBuffer(plain))
		}

		if len(bodyBytes) > 0 {
			var data map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &data); err == nil {
				for k, v := range data {
					params[k] = toString(v)
				}
			}
		}
	}

	return params
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case nil:
		return ""
	default:
		return ""
	}
}
