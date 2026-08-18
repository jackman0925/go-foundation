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
	// SlashDateLayout 是斜杠日期格式。
	SlashDateLayout = "2006/01/02"
	// SlashDateTimeLayout 是斜杠日期时间格式。
	SlashDateTimeLayout = "2006/01/02 15:04:05"
	// CompactDateLayout 是紧凑日期格式。
	CompactDateLayout = "20060102"
	// CompactDateTimeLayout 是紧凑日期时间格式。
	CompactDateTimeLayout = "20060102150405"
	// ShortCompactDateLayout 是短年份紧凑日期格式。
	ShortCompactDateLayout = "060102"
	// ShortCompactDateTimeLayout 是短年份紧凑日期时间格式。
	ShortCompactDateTimeLayout = "060102150405"
	// TimeLayout 是标准时间格式。
	TimeLayout = "15:04:05"
)

var parseLayouts = []string{
	DateTimeLayout,
	DateLayout,
	SlashDateTimeLayout,
	SlashDateLayout,
	CompactDateTimeLayout,
	CompactDateLayout,
	time.RFC3339,
	time.RFC3339Nano,
}

var weekdayCN = []string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}

// StartOfDay 返回输入时间所在日期的开始时间，并保留原时区。
func StartOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// EndOfDay 返回输入时间所在日期的最后一个纳秒，并保留原时区。
func EndOfDay(t time.Time) time.Time {
	return StartOfDay(t).AddDate(0, 0, 1).Add(-time.Nanosecond)
}

// StartOfWeek 返回输入时间所在自然周的开始时间，周一作为第一天。
func StartOfWeek(t time.Time) time.Time {
	return StartOfDay(t).AddDate(0, 0, -ISOWeekday(t)+1)
}

// EndOfWeek 返回输入时间所在自然周的结束时间，周日作为最后一天。
func EndOfWeek(t time.Time) time.Time {
	return StartOfWeek(t).AddDate(0, 0, 7).Add(-time.Nanosecond)
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

// FormatCompactDate 使用 CompactDateLayout 格式化时间。
func FormatCompactDate(t time.Time) string {
	return t.Format(CompactDateLayout)
}

// FormatCompactDateTime 使用 CompactDateTimeLayout 格式化时间。
func FormatCompactDateTime(t time.Time) string {
	return t.Format(CompactDateTimeLayout)
}

// FormatCompactDateTimeMilli 格式化为 yyyyMMddHHmmssSSS。
func FormatCompactDateTimeMilli(t time.Time) string {
	return t.Format(CompactDateTimeLayout) + fmt.Sprintf("%03d", t.Nanosecond()/int(time.Millisecond))
}

// FormatShortCompactDateTimeMilli 格式化为 yyMMddHHmmssSSS。
func FormatShortCompactDateTimeMilli(t time.Time) string {
	return t.Format(ShortCompactDateTimeLayout) + fmt.Sprintf("%03d", t.Nanosecond()/int(time.Millisecond))
}

// ParseTime 解析常见日期和日期时间格式。
func ParseTime(value string) (time.Time, error) {
	return ParseTimeInLocation(value, time.Local)
}

// ParseTimeInLocation 按指定时区解析常见日期和日期时间格式。
func ParseTimeInLocation(value string, location *time.Location) (time.Time, error) {
	if location == nil {
		location = time.Local
	}
	for _, layout := range parseLayouts {
		if parsed, err := time.ParseInLocation(layout, value, location); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("parse time %q: unsupported layout", value)
}

// NowUnixMilli 返回当前 Unix 毫秒时间戳。
func NowUnixMilli() int64 {
	return time.Now().UnixMilli()
}

// UnixMilli 返回时间对应的 Unix 毫秒时间戳。
func UnixMilli(t time.Time) int64 {
	return t.UnixMilli()
}

// FromUnixMilli 将 Unix 毫秒时间戳转换为时间。
func FromUnixMilli(millis int64) time.Time {
	return time.UnixMilli(millis)
}

// SecondsBetween 返回 second 相对 first 的秒数差值。
func SecondsBetween(first time.Time, second time.Time) int64 {
	return int64(second.Sub(first).Seconds())
}

// CompareDateString 按日期比较两个 DateLayout 字符串。
func CompareDateString(first string, second string) (int, error) {
	firstDate, err := time.ParseInLocation(DateLayout, first, time.Local)
	if err != nil {
		return 0, fmt.Errorf("parse first date %q: %w", first, err)
	}
	secondDate, err := time.ParseInLocation(DateLayout, second, time.Local)
	if err != nil {
		return 0, fmt.Errorf("parse second date %q: %w", second, err)
	}
	if firstDate.Before(secondDate) {
		return -1, nil
	}
	if firstDate.After(secondDate) {
		return 1, nil
	}
	return 0, nil
}

// Age 返回生日到当前时间的周岁。
func Age(birthday time.Time) int {
	return AgeAt(birthday, time.Now())
}

// AgeAt 返回生日到指定时间的周岁。
func AgeAt(birthday time.Time, now time.Time) int {
	if birthday.After(now) {
		return 0
	}

	age := now.Year() - birthday.Year()
	birthdayThisYear := time.Date(now.Year(), birthday.Month(), birthday.Day(), 0, 0, 0, 0, now.Location())
	if StartOfDay(now).Before(birthdayThisYear) {
		age--
	}
	if age < 0 {
		return 0
	}
	return age
}

// AddYears 给时间增加指定年数。
func AddYears(t time.Time, years int) time.Time {
	return t.AddDate(years, 0, 0)
}

// AddMonths 给时间增加指定月数。
func AddMonths(t time.Time, months int) time.Time {
	return t.AddDate(0, months, 0)
}

// AddDays 给时间增加指定天数。
func AddDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, days)
}

// AddHours 给时间增加指定小时数。
func AddHours(t time.Time, hours int) time.Time {
	return t.Add(time.Duration(hours) * time.Hour)
}

// AddSeconds 给时间增加指定秒数。
func AddSeconds(t time.Time, seconds int) time.Time {
	return t.Add(time.Duration(seconds) * time.Second)
}

// ChineseWeekday 返回中文星期。
func ChineseWeekday(t time.Time) string {
	return weekdayCN[int(t.Weekday())]
}

// ISOWeekday 返回 ISO 星期数字，周一为 1，周日为 7。
func ISOWeekday(t time.Time) int {
	weekday := int(t.Weekday())
	if weekday == 0 {
		return 7
	}
	return weekday
}

// ChineseWeekdayByISO 按 ISO 星期数字返回中文星期。
func ChineseWeekdayByISO(value int) string {
	if value < 1 || value > 7 {
		return ""
	}
	if value == 7 {
		return weekdayCN[0]
	}
	return weekdayCN[value]
}
