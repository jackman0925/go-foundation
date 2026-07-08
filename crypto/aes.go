package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// EncryptAESGCM encrypts plaintext with AES-GCM and returns base64(nonce|ciphertext).
func EncryptAESGCM(plaintext string, key []byte) (string, error) {
	gcm, err := newAESGCM(key)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	sealed := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	payload := make([]byte, 0, len(nonce)+len(sealed))
	payload = append(payload, nonce...)
	payload = append(payload, sealed...)

	return base64.StdEncoding.EncodeToString(payload), nil
}

// DecryptAESGCM decrypts base64(nonce|ciphertext) produced by EncryptAESGCM.
func DecryptAESGCM(ciphertext string, key []byte) (string, error) {
	gcm, err := newAESGCM(key)
	if err != nil {
		return "", err
	}

	payload, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	if len(payload) <= gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce := payload[:gcm.NonceSize()]
	sealed := payload[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func newAESGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
