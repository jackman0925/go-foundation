package timex

import (
	"fmt"
	"time"
)

const (
	// DateLayout is the standard date format used by this package.
	DateLayout = "2006-01-02"
	// DateTimeLayout is the standard date-time format used by this package.
	DateTimeLayout = "2006-01-02 15:04:05"
)

var parseLayouts = []string{
	DateTimeLayout,
	DateLayout,
	time.RFC3339,
	time.RFC3339Nano,
}

// StartOfDay returns the beginning of the input day in the same location.
func StartOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the final nanosecond of the input day in the same location.
func EndOfDay(t time.Time) time.Time {
	return StartOfDay(t).AddDate(0, 0, 1).Add(-time.Nanosecond)
}

// StartOfMonth returns the beginning of the input month in the same location.
func StartOfMonth(t time.Time) time.Time {
	year, month, _ := t.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns the final nanosecond of the input month in the same location.
func EndOfMonth(t time.Time) time.Time {
	return StartOfMonth(t).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

// FormatDate formats time using DateLayout.
func FormatDate(t time.Time) string {
	return t.Format(DateLayout)
}

// FormatDateTime formats time using DateTimeLayout.
func FormatDateTime(t time.Time) string {
	return t.Format(DateTimeLayout)
}

// Parse parses common date and date-time layouts.
func Parse(value string) (time.Time, error) {
	for _, layout := range parseLayouts {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("parse time %q: unsupported layout", value)
}
