package crypto

import "golang.org/x/crypto/bcrypt"

// HashPassword 使用 bcrypt.DefaultCost 对密码生成哈希。
func HashPassword(password string) (string, error) {
	content, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// CheckPassword 判断密码是否匹配 bcrypt 哈希。
func CheckPassword(password string, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
