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
	got, err := Parse("2026-07-08 09:10:11")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if got.Year() != 2026 || got.Month() != 7 || got.Day() != 8 || got.Hour() != 9 {
		t.Fatalf("unexpected parsed time: %s", got)
	}
}

func TestParseRFC3339AndRejectsInvalidValue(t *testing.T) {
	got, err := Parse("2026-07-08T09:10:11Z")
	if err != nil {
		t.Fatalf("Parse RFC3339 returned error: %v", err)
	}
	if got.Location() != time.UTC {
		t.Fatalf("expected UTC location, got %s", got.Location())
	}

	if _, err := Parse("not-a-time"); err == nil {
		t.Fatal("expected invalid time error")
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
