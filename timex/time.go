package timex

import (
	"fmt"
	"time"
)

const (
	// DateLayout 是本包使用的标准日期格式。
	DateLayout = "2006-01-02"
	// DateTimeLayout 是本包使用的标准日期时间格式。
	DateTimeLayout = "2006-01-02 15:04:05"
)

var parseLayouts = []string{
	DateTimeLayout,
	DateLayout,
	time.RFC3339,
	time.RFC3339Nano,
}

// StartOfDay 返回输入时间所在日期的开始时间，并保留原时区。
func StartOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// EndOfDay 返回输入时间所在日期的最后一个纳秒，并保留原时区。
func EndOfDay(t time.Time) time.Time {
	return StartOfDay(t).AddDate(0, 0, 1).Add(-time.Nanosecond)
}

// StartOfMonth 返回输入时间所在月份的开始时间，并保留原时区。
func StartOfMonth(t time.Time) time.Time {
	year, month, _ := t.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth 返回输入时间所在月份的最后一个纳秒，并保留原时区。
func EndOfMonth(t time.Time) time.Time {
	return StartOfMonth(t).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

// FormatDate 使用 DateLayout 格式化时间。
func FormatDate(t time.Time) string {
	return t.Format(DateLayout)
}

// FormatDateTime 使用 DateTimeLayout 格式化时间。
func FormatDateTime(t time.Time) string {
	return t.Format(DateTimeLayout)
}

// Parse 解析常见日期和日期时间格式。
func Parse(value string) (time.Time, error) {
	for _, layout := range parseLayouts {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("parse time %q: unsupported layout", value)
}
