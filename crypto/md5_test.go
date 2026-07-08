package crypto

import "testing"

func TestMD5Hex16(t *testing.T) {
	got := MD5Hex16("hello")
	if got != "bc4b2a76b9719d91" {
		t.Fatalf("unexpected 16-char md5: %s", got)
	}
}
