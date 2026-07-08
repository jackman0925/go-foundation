package crypto

import "golang.org/x/crypto/bcrypt"

// HashPassword hashes a password using bcrypt.DefaultCost.
func HashPassword(password string) (string, error) {
	content, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// CheckPassword reports whether password matches a bcrypt hash.
func CheckPassword(password string, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
