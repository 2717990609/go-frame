// Package signature 客户端签名生成（供 SDK、测试或前端对接参考）
package signature

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"time"
)

// GenerateWithNonce 生成带 timestamp、nonce 的签名，返回签名及需设置的 Header 值
// params 为业务参数（不含 timestamp、nonce）
func GenerateWithNonce(params map[string]string, secret string) (sign, timestamp, nonce string) {
	timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	nonce = generateNonce()
	p := make(map[string]string)
	for k, v := range params {
		p[k] = v
	}
	p["timestamp"] = timestamp
	p["nonce"] = nonce
	sign = Generate(p, secret)
	return
}

func generateNonce() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
