package stringx

import (
	"regexp"
	"strconv"
)

var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// IsNumeric reports whether value is a base-10 integer string.
func IsNumeric(value string) bool {
	if value == "" {
		return false
	}
	_, err := strconv.ParseInt(value, 10, 64)
	return err == nil
}

// IsEmail reports whether value looks like a valid email address.
func IsEmail(value string) bool {
	return emailPattern.MatchString(value)
}
