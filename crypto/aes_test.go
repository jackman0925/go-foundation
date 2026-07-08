package crypto

import "testing"

func TestEncryptAESGCMAndDecryptAESGCM(t *testing.T) {
	key := []byte("1234567890abcdef1234567890abcdef")

	ciphertext, err := EncryptAESGCM("hello", key)
	if err != nil {
		t.Fatalf("EncryptAESGCM returned error: %v", err)
	}
	if ciphertext == "" || ciphertext == "hello" {
		t.Fatalf("expected encrypted base64 text, got %q", ciphertext)
	}

	plaintext, err := DecryptAESGCM(ciphertext, key)
	if err != nil {
		t.Fatalf("DecryptAESGCM returned error: %v", err)
	}
	if plaintext != "hello" {
		t.Fatalf("expected hello, got %q", plaintext)
	}
}

func TestEncryptAESGCMUsesRandomNonce(t *testing.T) {
	key := []byte("1234567890abcdef1234567890abcdef")

	left, err := EncryptAESGCM("hello", key)
	if err != nil {
		t.Fatalf("EncryptAESGCM returned error: %v", err)
	}
	right, err := EncryptAESGCM("hello", key)
	if err != nil {
		t.Fatalf("EncryptAESGCM returned error: %v", err)
	}

	if left == right {
		t.Fatal("expected different ciphertexts for same plaintext")
	}
}

func TestDecryptAESGCMRejectsWrongKey(t *testing.T) {
	key := []byte("1234567890abcdef1234567890abcdef")
	wrongKey := []byte("abcdef1234567890abcdef1234567890")

	ciphertext, err := EncryptAESGCM("hello", key)
	if err != nil {
		t.Fatalf("EncryptAESGCM returned error: %v", err)
	}

	if _, err := DecryptAESGCM(ciphertext, wrongKey); err == nil {
		t.Fatal("expected wrong key error")
	}
}

func TestAESGCMRejectsInvalidInput(t *testing.T) {
	if _, err := EncryptAESGCM("hello", []byte("short")); err == nil {
		t.Fatal("expected invalid key error")
	}
	if _, err := DecryptAESGCM("not-base64", []byte("1234567890abcdef")); err == nil {
		t.Fatal("expected invalid ciphertext error")
	}
}
