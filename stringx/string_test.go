package stringx

import "testing"

func TestIsBlankTrimsWhitespace(t *testing.T) {
	if !IsBlank(" \t\n") {
		t.Fatal("expected whitespace string to be blank")
	}
	if IsBlank(" demo ") {
		t.Fatal("expected non-empty string to not be blank")
	}
}

func TestTruncateHandlesRuneBoundaries(t *testing.T) {
	got := Truncate("你好世界", 2)
	if got != "你好" {
		t.Fatalf("expected rune-safe truncate, got %q", got)
	}
}

func TestMaskMobile(t *testing.T) {
	got := MaskMobile("13800138000")
	if got != "138****8000" {
		t.Fatalf("unexpected masked mobile: %q", got)
	}
}

func TestRandomStringLength(t *testing.T) {
	got, err := RandomString(16)
	if err != nil {
		t.Fatalf("RandomString returned error: %v", err)
	}
	if len(got) != 16 {
		t.Fatalf("expected length 16, got %d", len(got))
	}
}
