// Package signature 请求签名与验签，符合规范「任何外部回调必须验签」
package signature

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Generate 生成签名，params 需包含 timestamp 和 nonce
// 算法：参数按键名排序拼接 key1=v1&key2=v2&key=secret，SHA256 后转大写
func Generate(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" || k == "signature" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		v := params[k]
		if b.Len() > 0 {
			b.WriteString("&")
		}
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(v)
	}
	b.WriteString("&key=")
	b.WriteString(secret)
	hash := sha256.Sum256([]byte(b.String()))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// Verify 验签，校验签名一致性
func Verify(params map[string]string, sign string, secret string) bool {
	expected := Generate(params, secret)
	return strings.EqualFold(sign, expected)
}

// DefaultTimeWindow 默认时间窗口 5 分钟
const DefaultTimeWindow = 5 * time.Minute

// ValidateTimestamp 校验时间戳是否在允许窗口内
func ValidateTimestamp(timestampStr string, window time.Duration) error {
	if window <= 0 {
		window = DefaultTimeWindow
	}
	ts, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return fmt.Errorf("时间戳格式错误")
	}
	t := time.Unix(ts, 0)
	now := time.Now()
	if now.Sub(t) > window || t.Sub(now) > window {
		return fmt.Errorf("请求已过期，请检查设备时间")
	}
	return nil
}
