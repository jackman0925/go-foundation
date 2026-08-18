package timex

import (
	"testing"
	"time"
)

func TestStartAndEndOfDayPreserveLocation(t *testing.T) {
	loc := time.FixedZone("UTC+8", 8*60*60)
	input := time.Date(2026, 7, 8, 15, 30, 45, 123, loc)

	start := StartOfDay(input)
	end := EndOfDay(input)

	if !start.Equal(time.Date(2026, 7, 8, 0, 0, 0, 0, loc)) {
		t.Fatalf("unexpected start of day: %s", start)
	}
	if !end.Equal(time.Date(2026, 7, 8, 23, 59, 59, int(time.Second-time.Nanosecond), loc)) {
		t.Fatalf("unexpected end of day: %s", end)
	}
}

func TestStartAndEndOfMonth(t *testing.T) {
	input := time.Date(2024, 2, 20, 12, 0, 0, 0, time.UTC)

	if got := StartOfMonth(input); !got.Equal(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected start of month: %s", got)
	}
	if got := EndOfMonth(input); !got.Equal(time.Date(2024, 2, 29, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)) {
		t.Fatalf("unexpected end of month: %s", got)
	}
}

func TestParseCommonLayouts(t *testing.T) {
	got, err := ParseTime("2026-07-08 09:10:11")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if got.Year() != 2026 || got.Month() != 7 || got.Day() != 8 || got.Hour() != 9 {
		t.Fatalf("unexpected parsed time: %s", got)
	}
}

func TestParseSupportsAdditionalLayouts(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{value: "2026/07/08 09:10:11", want: "2026-07-08 09:10:11"},
		{value: "20260708091011", want: "2026-07-08 09:10:11"},
		{value: "2026/07/08", want: "2026-07-08 00:00:00"},
		{value: "20260708", want: "2026-07-08 00:00:00"},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := ParseTime(tt.value)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			if FormatDateTime(got) != tt.want {
				t.Fatalf("expected %s, got %s", tt.want, FormatDateTime(got))
			}
		})
	}
}

func TestParseTimeInLocationPreservesLocationForLocalLayouts(t *testing.T) {
	loc := time.FixedZone("UTC+8", 8*60*60)
	got, err := ParseTimeInLocation("2026-07-08 09:10:11", loc)
	if err != nil {
		t.Fatalf("ParseTimeInLocation returned error: %v", err)
	}

	if got.Location() != loc {
		t.Fatalf("expected custom location, got %s", got.Location())
	}
	if got.Hour() != 9 {
		t.Fatalf("expected hour 9, got %d", got.Hour())
	}

	got, err = ParseTimeInLocation("2026-07-08 09:10:11", nil)
	if err != nil {
		t.Fatalf("ParseTimeInLocation nil location returned error: %v", err)
	}
	if got.Location() != time.Local {
		t.Fatalf("expected local location, got %s", got.Location())
	}
}

func TestParseRFC3339AndRejectsInvalidValue(t *testing.T) {
	got, err := ParseTime("2026-07-08T09:10:11Z")
	if err != nil {
		t.Fatalf("Parse RFC3339 returned error: %v", err)
	}
	if got.Location() != time.UTC {
		t.Fatalf("expected UTC location, got %s", got.Location())
	}

	if _, err := ParseTime("not-a-time"); err == nil {
		t.Fatal("expected invalid time error")
	}
}

func TestUnixMilliHelpers(t *testing.T) {
	input := time.Date(2026, 7, 8, 9, 10, 11, 123*int(time.Millisecond), time.UTC)

	nowMillis := NowUnixMilli()
	if nowMillis <= 0 {
		t.Fatalf("expected positive current millis, got %d", nowMillis)
	}

	millis := UnixMilli(input)
	if millis != 1783501811123 {
		t.Fatalf("unexpected millis: %d", millis)
	}

	got := FromUnixMilli(millis)
	if !got.Equal(input) {
		t.Fatalf("expected %s, got %s", input, got)
	}
}

func TestDateTimeAddHelpers(t *testing.T) {
	input := time.Date(2026, 1, 31, 10, 0, 0, 0, time.UTC)

	if got := AddYears(input, 1); got.Year() != 2027 {
		t.Fatalf("unexpected AddYears result: %s", got)
	}
	if got := AddMonths(input, 1); got.Month() != time.March || got.Day() != 3 {
		t.Fatalf("unexpected AddMonths result: %s", got)
	}
	if got := AddDays(input, -1); got.Day() != 30 {
		t.Fatalf("unexpected AddDays result: %s", got)
	}
	if got := AddHours(input, 2); got.Hour() != 12 {
		t.Fatalf("unexpected AddHours result: %s", got)
	}
	if got := AddSeconds(input, 30); got.Second() != 30 {
		t.Fatalf("unexpected AddSeconds result: %s", got)
	}
}

func TestSecondsBetweenAndCompareDateString(t *testing.T) {
	first := time.Date(2026, 7, 8, 9, 0, 0, 0, time.UTC)
	second := time.Date(2026, 7, 8, 9, 1, 30, 0, time.UTC)

	if got := SecondsBetween(first, second); got != 90 {
		t.Fatalf("expected 90 seconds, got %d", got)
	}
	if got := SecondsBetween(second, first); got != -90 {
		t.Fatalf("expected -90 seconds, got %d", got)
	}

	result, err := CompareDateString("2026-07-08", "2026-07-09")
	if err != nil {
		t.Fatalf("CompareDateString returned error: %v", err)
	}
	if result != -1 {
		t.Fatalf("expected -1, got %d", result)
	}

	result, err = CompareDateString("2026-07-09", "2026-07-08")
	if err != nil {
		t.Fatalf("CompareDateString returned error: %v", err)
	}
	if result != 1 {
		t.Fatalf("expected 1, got %d", result)
	}

	result, err = CompareDateString("2026-07-08", "2026-07-08")
	if err != nil {
		t.Fatalf("CompareDateString returned error: %v", err)
	}
	if result != 0 {
		t.Fatalf("expected 0, got %d", result)
	}
}

func TestCompareDateRejectsInvalidDate(t *testing.T) {
	if _, err := CompareDateString("bad", "2026-07-08"); err == nil {
		t.Fatal("expected invalid first date error")
	}
	if _, err := CompareDateString("2026-07-08", "bad"); err == nil {
		t.Fatal("expected invalid second date error")
	}
}

func TestAgeAt(t *testing.T) {
	now := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)

	if got := Age(time.Now().AddDate(-1, 0, 0)); got < 0 {
		t.Fatalf("age should not be negative, got %d", got)
	}
	if got := AgeAt(time.Date(2000, 7, 8, 0, 0, 0, 0, time.UTC), now); got != 26 {
		t.Fatalf("expected age 26, got %d", got)
	}
	if got := AgeAt(time.Date(2000, 7, 9, 0, 0, 0, 0, time.UTC), now); got != 25 {
		t.Fatalf("expected age 25, got %d", got)
	}
	if got := AgeAt(time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), now); got != 0 {
		t.Fatalf("future birthday should return 0, got %d", got)
	}
}

func TestFormatCompactDateTimeAndMillis(t *testing.T) {
	input := time.Date(2026, 7, 8, 9, 10, 11, 123*int(time.Millisecond), time.UTC)

	if got := FormatCompactDate(input); got != "20260708" {
		t.Fatalf("unexpected compact date: %s", got)
	}
	if got := FormatCompactDateTime(input); got != "20260708091011" {
		t.Fatalf("unexpected compact datetime: %s", got)
	}
	if got := FormatCompactDateTimeMilli(input); got != "20260708091011123" {
		t.Fatalf("unexpected datetime millis: %s", got)
	}
	if got := FormatShortCompactDateTimeMilli(input); got != "260708091011123" {
		t.Fatalf("unexpected short datetime millis: %s", got)
	}
}

func TestWeekHelpers(t *testing.T) {
	monday := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	sunday := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)

	if got := ChineseWeekday(monday); got != "星期一" {
		t.Fatalf("expected 星期一, got %s", got)
	}
	if got := ISOWeekday(monday); got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
	if got := ChineseWeekday(sunday); got != "星期日" {
		t.Fatalf("expected 星期日, got %s", got)
	}
	if got := ISOWeekday(sunday); got != 7 {
		t.Fatalf("expected 7, got %d", got)
	}
	if got := ChineseWeekdayByISO(8); got != "" {
		t.Fatalf("expected empty weekday, got %s", got)
	}
	if got := ChineseWeekdayByISO(1); got != "星期一" {
		t.Fatalf("expected 星期一, got %s", got)
	}
	if got := ChineseWeekdayByISO(7); got != "星期日" {
		t.Fatalf("expected 星期日, got %s", got)
	}
}

func TestStartAndEndOfWeek(t *testing.T) {
	loc := time.FixedZone("UTC+8", 8*60*60)
	input := time.Date(2026, 7, 8, 9, 10, 11, 0, loc)

	start := StartOfWeek(input)
	end := EndOfWeek(input)

	if !start.Equal(time.Date(2026, 7, 6, 0, 0, 0, 0, loc)) {
		t.Fatalf("unexpected start of week: %s", start)
	}
	if !end.Equal(time.Date(2026, 7, 12, 23, 59, 59, int(time.Second-time.Nanosecond), loc)) {
		t.Fatalf("unexpected end of week: %s", end)
	}
}

func TestFormatDateAndDateTime(t *testing.T) {
	input := time.Date(2026, 7, 8, 9, 10, 11, 0, time.UTC)

	if got := FormatDate(input); got != "2026-07-08" {
		t.Fatalf("unexpected date: %q", got)
	}
	if got := FormatDateTime(input); got != "2026-07-08 09:10:11" {
		t.Fatalf("unexpected date time: %q", got)
	}
}
