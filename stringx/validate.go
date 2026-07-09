package stringx

import (
	"regexp"
	"strconv"
)

var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// IsNumeric 判断字符串是否为十进制整数。
func IsNumeric(value string) bool {
	if value == "" {
		return false
	}
	_, err := strconv.ParseInt(value, 10, 64)
	return err == nil
}

// IsEmail 判断字符串是否符合常见邮箱格式。
func IsEmail(value string) bool {
	return emailPattern.MatchString(value)
}
