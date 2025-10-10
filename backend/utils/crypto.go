package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GenerateRandomToken generates a cryptographically secure random token
func GenerateRandomToken() string {
	bytes := make([]byte, 32) // 32 bytes = 64 hex characters
	if _, err := rand.Read(bytes); err != nil {
		panic(err) // This should never happen in normal operation
	}
	return hex.EncodeToString(bytes)
}

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}

	for i := range bytes {
		bytes[i] = charset[bytes[i]%byte(len(charset))]
	}

	return string(bytes)
}

// HashPassword hashes a password using SHA-256
func HashPassword(password string) (string, error) {
	hash := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", hash), nil
}
