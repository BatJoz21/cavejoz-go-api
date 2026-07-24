package utils

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)

	return string(bytes), err
}

func HashRefreshToken(refreshToken string) string {
	hashed := sha256.Sum256([]byte(refreshToken))

	return hex.EncodeToString(hashed[:])
}

func CheckPasswordHash(rawPw, hashedPw string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPw), []byte(rawPw))

	return err == nil
}
