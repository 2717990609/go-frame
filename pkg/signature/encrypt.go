// Package signature 参数加解密，适用于敏感接口
package signature

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

const gcmNonceSize = 12

// EncryptAESGCM 使用 AES-256-GCM 加密，返回 Base64
// key 需 32 字节，不足则用 0 补齐或截断
func EncryptAESGCM(plaintext []byte, key []byte) (string, error) {
	key = normalizeKey(key, 32)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAESGCM 解密 Base64 密文
func DecryptAESGCM(ciphertextB64 string, key []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, errors.New("密文格式错误")
	}
	key = normalizeKey(key, 32)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("密文长度不足")
	}
	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func normalizeKey(key []byte, size int) []byte {
	if len(key) >= size {
		return key[:size]
	}
	padded := make([]byte, size)
	copy(padded, key)
	return padded
}

// DecryptBody 若请求体为加密格式，解密后返回明文；否则返回原 Body
func DecryptBody(body []byte, encrypted bool, key []byte) ([]byte, error) {
	if !encrypted || len(body) == 0 {
		return body, nil
	}
	plain, err := DecryptAESGCM(string(bytes.TrimSpace(body)), key)
	if err != nil {
		return nil, err
	}
	return plain, nil
}
