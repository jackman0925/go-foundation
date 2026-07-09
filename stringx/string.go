package stringx

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

const randomAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// IsBlank 判断字符串去除空白后是否为空。
func IsBlank(value string) bool {
	return strings.TrimSpace(value) == ""
}

// Truncate 按 rune 截断字符串，避免切断多字节字符。
func Truncate(value string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	return string(runes[:maxRunes])
}

// MaskMobile 对 11 位手机号中间四位做脱敏处理。
func MaskMobile(value string) string {
	runes := []rune(value)
	if len(runes) != 11 {
		return value
	}
	return string(runes[:3]) + "****" + string(runes[7:])
}

// RandomString 生成加密安全的随机字母数字字符串。
func RandomString(length int) (string, error) {
	if length < 0 {
		return "", fmt.Errorf("length must be non-negative")
	}

	result := make([]byte, length)
	for i := range result {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(randomAlphabet))))
		if err != nil {
			return "", err
		}
		result[i] = randomAlphabet[index.Int64()]
	}
	return string(result), nil
}
