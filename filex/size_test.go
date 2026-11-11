package filex

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"
)

func TestSizeDisplayFixtures(t *testing.T) {
	data, err := os.ReadFile("testdata/size_display.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct{ Bytes, Decimal, Binary string }
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("empty fixtures")
	}
	for _, tc := range cases {
		t.Run(tc.Bytes, func(t *testing.T) {
			value, err := strconv.ParseInt(tc.Bytes, 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			if got := FormatDecimalSize(value); got != tc.Decimal {
				t.Errorf("decimal: got %q, want %q", got, tc.Decimal)
			}
			if got := FormatBinarySize(value); got != tc.Binary {
				t.Errorf("binary: got %q, want %q", got, tc.Binary)
			}
		})
	}
}

func TestSizeUnitBoundaries(t *testing.T) {
	for _, mode := range []struct {
		base   int64
		units  []string
		format func(int64) string
	}{
		{1000, []string{"B", "kB", "MB", "GB", "TB", "PB"}, FormatDecimalSize},
		{1024, []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB"}, FormatBinarySize},
	} {
		boundary := int64(1)
		for unit := 1; unit < len(mode.units); unit++ {
			boundary *= mode.base
			for _, delta := range []int64{-1, 0, 1} {
				value := boundary + delta
				want := "1 " + mode.units[unit]
				if unit == 1 && delta == -1 {
					want = strconv.FormatInt(value, 10) + " B"
				}
				if got := mode.format(value); got != want {
					t.Errorf("%d: got %q, want %q", value, got, want)
				}
				if got := mode.format(-value); got != "-"+want {
					t.Errorf("%d: got %q, want -%s", -value, got, want)
				}
			}
		}
	}
}

func TestSizeRoundingTies(t *testing.T) {
	for _, tc := range []struct {
		value int64
		want  string
	}{
		{1004999999, "1 GB"}, {1005000000, "1.01 GB"}, {1005000001, "1.01 GB"},
		{999994999999, "999.99 GB"}, {999995000000, "1 TB"},
		{-1050, "-1.1 kB"},
	} {
		if got := FormatDecimalSize(tc.value); got != tc.want {
			t.Errorf("%d: got %s, want %s", tc.value, got, tc.want)
		}
	}
	if got := FormatBinarySize(1152); got != "1.1 KiB" {
		t.Fatal(got)
	}
	if got := FormatBinarySize(1280); got != "1.3 KiB" {
		t.Fatal(got)
	}
}
