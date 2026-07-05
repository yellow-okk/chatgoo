package hash

import "golang.org/x/crypto/bcrypt"

// Bcrypt generates a bcrypt hash of the password.
func Bcrypt(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

// VerifyBcrypt checks if a password matches a bcrypt hash.
func VerifyBcrypt(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
