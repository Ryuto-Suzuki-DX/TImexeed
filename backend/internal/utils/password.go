package utils

import "golang.org/x/crypto/bcrypt"

/*
 * パスワードをハッシュ化する
 */
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

/*
 * パスワードとハッシュ化済みパスワードを比較する
 */
func CheckPassword(password string, passwordHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	return err == nil
}
