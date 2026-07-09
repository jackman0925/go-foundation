package crypto

import (
	"fmt"
	"sort"
	"strings"
)

// MapChecksumMD5 为字符串键 map 生成稳定的 MD5 校验值。
func MapChecksumMD5(data map[string]any) string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var builder strings.Builder
	for _, key := range keys {
		builder.WriteString(key)
		builder.WriteByte('=')
		builder.WriteString(fmt.Sprintf("%#v", data[key]))
		builder.WriteByte('\n')
	}

	return MD5Hex(builder.String())
}
