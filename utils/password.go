package utils

import (
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash compares a password with a hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// IsStrongPassword checks if a password meets strength requirements
func IsStrongPassword(password string) bool {
	// At least 8 characters
	if len(password) < 8 {
		return false
	}

	// Contains at least one uppercase letter
	uppercase := regexp.MustCompile(`[A-Z]`)
	if !uppercase.MatchString(password) {
		return false
	}

	// Contains at least one lowercase letter
	lowercase := regexp.MustCompile(`[a-z]`)
	if !lowercase.MatchString(password) {
		return false
	}

	// Contains at least one digit
	digit := regexp.MustCompile(`[0-9]`)
	if !digit.MatchString(password) {
		return false
	}

	// Contains at least one special character
	special := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`)
	if !special.MatchString(password) {
		return false
	}

	return true
}
