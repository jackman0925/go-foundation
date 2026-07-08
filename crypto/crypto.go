package crypto

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

const digits = "0123456789"

// MD5Hex returns the lowercase hex-encoded MD5 digest of value.
func MD5Hex(value string) string {
	sum := md5.Sum([]byte(value))
	return hex.EncodeToString(sum[:])
}

// MD5Hex16 returns the middle 16 characters of the lowercase MD5 digest.
func MD5Hex16(value string) string {
	return MD5Hex(value)[8:24]
}

// SHA256Hex returns the lowercase hex-encoded SHA256 digest of value.
func SHA256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

// HMACSHA256Hex returns the lowercase hex-encoded HMAC-SHA256 signature.
func HMACSHA256Hex(secret string, value string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

// RandomDigits returns a cryptographically random numeric string.
func RandomDigits(length int) (string, error) {
	if length < 0 {
		return "", fmt.Errorf("length must be non-negative")
	}

	result := make([]byte, length)
	for i := range result {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		result[i] = digits[index.Int64()]
	}
	return string(result), nil
}
