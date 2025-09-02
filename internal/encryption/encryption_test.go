package encryption

import (
	"os"
	"testing"
)

func TestEncryptionDecryption(t *testing.T) {
	testData := []byte("hello world its me your best friend hahah what fun we will have1")
	password := "testpassword123"

	// Test encryption
	encrypted, err := EncryptData(testData, password)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Test decryption
	decrypted, err := DecryptData(encrypted, password)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	// Verify data matches
	if string(decrypted) != string(testData) {
		t.Errorf("Decrypted data doesn't match original. Got: %s, Want: %s",
			string(decrypted), string(testData))
	}
}

func TestFileEncryption(t *testing.T) {
	// Create test file
	testContent := "hello world its me your best friend hahah what fun we will have1"
	testFile := "test_hello.txt"

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	password := "userpassword123"

	// Test file encryption
	encryptedData, err := EncryptFile(testFile, password)
	if err != nil {
		t.Fatalf("File encryption failed: %v", err)
	}

	// Test file decryption
	decryptedData, err := DecryptData(encryptedData, password)
	if err != nil {
		t.Fatalf("File decryption failed: %v", err)
	}

	if string(decryptedData) != testContent {
		t.Errorf("Decrypted content doesn't match. Got: %s, Want: %s",
			string(decryptedData), testContent)
	}
}
