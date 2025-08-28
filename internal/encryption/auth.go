package encryption

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Encryptor defines the interface for password operations
//
//go:generate mockgen -package mocks -source auth.go -destination ./mocks/mock_encryptor.go PasswordEncryptor
type PasswordEncryptor interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) error
}

// DefaultEncryptor implements Encryptor using the package functions
type DefaultPaswordEncryptor struct{}

// NewPaswordEncryptor builds new DefaultPaswordEncryptor
func NewPaswordEncryptor() *DefaultPaswordEncryptor {
	return &DefaultPaswordEncryptor{}
}

func (d *DefaultPaswordEncryptor) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

func (d *DefaultPaswordEncryptor) CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
