package crypto

import "testing"

func TestMD5Hex(t *testing.T) {
	got := MD5Hex("hello")
	if got != "5d41402abc4b2a76b9719d911017c592" {
		t.Fatalf("unexpected md5: %s", got)
	}
}

func TestSHA256Hex(t *testing.T) {
	got := SHA256Hex("hello")
	if got != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Fatalf("unexpected sha256: %s", got)
	}
}

func TestHMACSHA256Hex(t *testing.T) {
	got := HMACSHA256Hex("secret", "hello")
	if got != "88aab3ede8d3adf94d26ab90d3bafd4a2083070c3bcce9c014ee04a443847c0b" {
		t.Fatalf("unexpected hmac: %s", got)
	}
}

func TestRandomDigitsLength(t *testing.T) {
	got, err := RandomDigits(8)
	if err != nil {
		t.Fatalf("RandomDigits returned error: %v", err)
	}
	if len(got) != 8 {
		t.Fatalf("expected length 8, got %d", len(got))
	}
	for _, r := range got {
		if r < '0' || r > '9' {
			t.Fatalf("expected only digits, got %q in %q", r, got)
		}
	}
}
