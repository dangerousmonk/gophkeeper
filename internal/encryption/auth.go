package encryption

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

func CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// Removes sensitive data from error e.g password
func SanitizeError(err error) string {
	msg := err.Error()
	if strings.Contains(msg, "password") {
		msg = strings.ReplaceAll(msg, "password", "***")
	}

	if strings.Contains(msg, "auth") {
		msg = strings.ReplaceAll(msg, "auth", "***")
	}
	return msg
}
