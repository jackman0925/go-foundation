package stringx

import "testing"

func TestIsNumeric(t *testing.T) {
	if IsNumeric("") {
		t.Fatal("expected empty string to be non-numeric")
	}
	if !IsNumeric("12345") {
		t.Fatal("expected numeric string")
	}
	if IsNumeric("12.3") {
		t.Fatal("expected decimal string to be non-numeric integer")
	}
	if IsNumeric("abc") {
		t.Fatal("expected non-numeric string")
	}
}

func TestIsEmail(t *testing.T) {
	if !IsEmail("user@example.com") {
		t.Fatal("expected valid email")
	}
	if IsEmail("bad-email") {
		t.Fatal("expected invalid email")
	}
}
