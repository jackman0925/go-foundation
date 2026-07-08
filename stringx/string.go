package stringx

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

const randomAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// IsBlank reports whether value is empty after trimming whitespace.
func IsBlank(value string) bool {
	return strings.TrimSpace(value) == ""
}

// Truncate returns at most maxRunes runes without splitting multi-byte characters.
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

// MaskMobile masks the middle four digits of an 11-digit mobile number.
func MaskMobile(value string) string {
	runes := []rune(value)
	if len(runes) != 11 {
		return value
	}
	return string(runes[:3]) + "****" + string(runes[7:])
}

// RandomString returns a cryptographically random alphanumeric string.
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
