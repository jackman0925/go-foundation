package stringx

import (
	"strings"
	"unicode"
)

// CamelToSnake converts camelCase or PascalCase into snake_case.
func CamelToSnake(value string) string {
	if value == "" {
		return ""
	}

	var builder strings.Builder
	runes := []rune(value)
	for i, current := range runes {
		if unicode.IsUpper(current) {
			if i > 0 {
				prev := runes[i-1]
				nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
				if unicode.IsLower(prev) || (unicode.IsUpper(prev) && nextIsLower) {
					builder.WriteRune('_')
				}
			}
			builder.WriteRune(unicode.ToLower(current))
			continue
		}
		builder.WriteRune(current)
	}

	return builder.String()
}
