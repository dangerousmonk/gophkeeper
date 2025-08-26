package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

// Encryption constants
const (
	saltSize         = 16
	keySize          = 32 // AES-256
	pbkdf2Iterations = 4096
)

// keyFromPassword creates an encryption key from the user's password
func keyFromPassword(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, pbkdf2Iterations, keySize, sha256.New)
}

// EncryptData encrypts data using AES-GCM
func EncryptData(data []byte, password string) ([]byte, error) {
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	key := keyFromPassword(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nil, nonce, data, nil)

	// [salt:16][nonce:12][ciphertext:variable]
	encryptedData := make([]byte, saltSize+len(nonce)+len(ciphertext))
	copy(encryptedData[0:saltSize], salt)
	copy(encryptedData[saltSize:saltSize+gcm.NonceSize()], nonce)
	copy(encryptedData[saltSize+gcm.NonceSize():], ciphertext)

	return encryptedData, nil
}

// DecryptData decrypts data by using user password and stored encryptedData
func DecryptData(encryptedData []byte, password string) ([]byte, error) {
	if len(encryptedData) < saltSize {
		return nil, fmt.Errorf("invalid encrypted data: too short for salt")
	}
	salt := encryptedData[:saltSize]
	remainingData := encryptedData[saltSize:]

	key := keyFromPassword(password, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(remainingData) < nonceSize {
		return nil, fmt.Errorf("invalid encrypted data: too short for nonce")
	}

	nonce := remainingData[:nonceSize]
	ciphertext := remainingData[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// EncryptFile encrypts file by using user password
func EncryptFile(path string, password string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	encryptedData, err := EncryptData(content, password)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt file content: %w", err)
	}

	return encryptedData, nil
}
