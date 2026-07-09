package stringx

import (
	"strconv"
	"strings"
)

// JoinInt64 使用指定分隔符拼接 int64 切片。
func JoinInt64(values []int64, separator string) string {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = strconv.FormatInt(value, 10)
	}
	return strings.Join(parts, separator)
}
