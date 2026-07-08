package stringx

import (
	"regexp"
	"strconv"
)

var versionSeparator = regexp.MustCompile(`[._]`)

// CompareVersion compares dotted or underscored numeric versions.
func CompareVersion(left string, right string) int {
	if left == right {
		return 0
	}

	leftParts := versionSeparator.Split(left, -1)
	rightParts := versionSeparator.Split(right, -1)
	maxLen := len(leftParts)
	if len(rightParts) > maxLen {
		maxLen = len(rightParts)
	}

	for i := 0; i < maxLen; i++ {
		leftValue := versionPartAt(leftParts, i)
		rightValue := versionPartAt(rightParts, i)
		if leftValue > rightValue {
			return 1
		}
		if leftValue < rightValue {
			return -1
		}
	}

	return 0
}

func versionPartAt(parts []string, index int) int64 {
	if index >= len(parts) {
		return 0
	}

	value, err := strconv.ParseInt(parts[index], 10, 64)
	if err != nil {
		return 0
	}
	return value
}
