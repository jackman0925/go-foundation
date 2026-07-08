package crypto

import "testing"

func TestHashPasswordAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("secret")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" || hash == "secret" {
		t.Fatalf("expected non-empty hash different from password, got %q", hash)
	}
	if !CheckPassword("secret", hash) {
		t.Fatal("expected password to match hash")
	}
	if CheckPassword("wrong", hash) {
		t.Fatal("expected wrong password to fail")
	}
}
