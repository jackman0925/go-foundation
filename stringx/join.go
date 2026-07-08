package stringx

import (
	"strconv"
	"strings"
)

// JoinInt64 joins int64 values with separator.
func JoinInt64(values []int64, separator string) string {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = strconv.FormatInt(value, 10)
	}
	return strings.Join(parts, separator)
}
