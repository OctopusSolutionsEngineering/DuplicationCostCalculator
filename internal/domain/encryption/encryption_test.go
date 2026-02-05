package encryption

import (
	"strings"
	"testing"
)

func GetMockKey() string {
	return "0123456789abcdef0123456789abcdef" // 32 bytes for AES-256
}

func TestEncryptString_Success(t *testing.T) {
	// Arrange
	plainText := "Hello, World!"

	// Act
	encrypted, err := EncryptStringWrapper(plainText, GetMockKey)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if encrypted == "" {
		t.Error("Expected encrypted text to be non-empty")
	}

	if encrypted == plainText {
		t.Error("Encrypted text should not match plain text")
	}
}

func TestEncryptString_EmptyString(t *testing.T) {
	// Arrange
	plainText := ""

	// Act
	encrypted, err := EncryptStringWrapper(plainText, GetMockKey)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if encrypted == "" {
		t.Error("Expected encrypted text to be non-empty even for empty input")
	}
}

func TestEncryptString_LongString(t *testing.T) {
	// Arrange
	plainText := strings.Repeat("A", 10000)

	// Act
	encrypted, err := EncryptStringWrapper(plainText, GetMockKey)

	// Assert
	if err != nil {
		t.Errorf("Expected no error for long string, got %v", err)
	}

	if encrypted == "" {
		t.Error("Expected encrypted text to be non-empty")
	}
}

func TestEncryptString_SpecialCharacters(t *testing.T) {
	// Arrange
	plainText := "Special chars: !@#$%^&*(){}[]<>?,./;':\"\\|`~"

	// Act
	encrypted, err := EncryptStringWrapper(plainText, GetMockKey)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if encrypted == "" {
		t.Error("Expected encrypted text to be non-empty")
	}
}

func TestEncryptString_UnicodeFCharacters(t *testing.T) {
	// Arrange
	plainText := "Unicode: 你好世界 🌍 émojis 🎉"

	// Act
	encrypted, err := EncryptStringWrapper(plainText, GetMockKey)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if encrypted == "" {
		t.Error("Expected encrypted text to be non-empty")
	}
}

func TestDecryptString_Success(t *testing.T) {
	// Arrange
	plainText := "Hello, World!"
	encrypted, err := EncryptStringWrapper(plainText, GetMockKey)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Act
	decrypted, err := DecryptStringWrapper(encrypted, GetMockKey)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if decrypted != plainText {
		t.Errorf("Expected decrypted text to be '%s', got '%s'", plainText, decrypted)
	}
}

func TestDecryptString_EmptyString(t *testing.T) {
	// Arrange
	plainText := ""
	encrypted, err := EncryptStringWrapper(plainText, GetMockKey)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Act
	decrypted, err := DecryptStringWrapper(encrypted, GetMockKey)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if decrypted != plainText {
		t.Errorf("Expected decrypted text to be empty, got '%s'", decrypted)
	}
}

func TestDecryptString_InvalidBase64(t *testing.T) {
	// Arrange
	invalidEncrypted := "not-valid-base64!@#$%"

	// Act
	_, err := DecryptStringWrapper(invalidEncrypted, GetMockKey)

	// Assert
	if err == nil {
		t.Error("Expected error for invalid base64, got nil")
	}
}

func TestDecryptString_CorruptedData(t *testing.T) {
	// Arrange
	plainText := "Test data"
	encrypted, err := EncryptStringWrapper(plainText, GetMockKey)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Corrupt the encrypted data by changing a character
	corrupted := "A" + encrypted[1:]

	// Act
	_, err = DecryptStringWrapper(corrupted, GetMockKey)

	// Assert
	if err == nil {
		t.Error("Expected error for corrupted ciphertext, got nil")
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	// Arrange
	testCases := []string{
		"Simple text",
		"",
		"Special chars: !@#$%^&*()",
		"Unicode: 你好世界 🌍",
		strings.Repeat("Long text ", 1000),
		"Multi\nLine\nText",
		"Tabs\tand\tspaces",
	}

	for _, plainText := range testCases {
		t.Run(plainText, func(t *testing.T) {
			// Act
			encrypted, err := EncryptStringWrapper(plainText, GetMockKey)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			decrypted, err := DecryptStringWrapper(encrypted, GetMockKey)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			// Assert
			if decrypted != plainText {
				t.Errorf("Round trip failed. Expected '%s', got '%s'", plainText, decrypted)
			}
		})
	}
}

func TestEncryptString_DifferentOutputs(t *testing.T) {
	// Arrange
	plainText := "Same input"

	// Act - Encrypt the same text twice
	encrypted1, err1 := EncryptStringWrapper(plainText, GetMockKey)
	encrypted2, err2 := EncryptStringWrapper(plainText, GetMockKey)

	// Assert
	if err1 != nil || err2 != nil {
		t.Fatalf("Expected no errors, got %v and %v", err1, err2)
	}

	// Encrypted outputs should be different due to random nonce
	if encrypted1 == encrypted2 {
		t.Error("Expected different encrypted outputs for same input (due to random nonce)")
	}

	// But both should decrypt to the same plaintext
	decrypted1, _ := DecryptStringWrapper(encrypted1, GetMockKey)
	decrypted2, _ := DecryptStringWrapper(encrypted2, GetMockKey)

	if decrypted1 != plainText || decrypted2 != plainText {
		t.Error("Both encrypted versions should decrypt to the original plaintext")
	}
}

func TestDecryptString_WrongKey(t *testing.T) {
	// Arrange
	plainText := "Secret message"
	correctKey := "0123456789abcdef0123456789abcdef" // 32 bytes for AES-256
	wrongKey := "fedcba9876543210fedcba9876543210"   // Different 32-byte key

	// Encrypt with correct key
	encrypted, err := EncryptStringWrapper(plainText, func() string { return correctKey })
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Act - Try to decrypt with wrong key
	_, err = DecryptStringWrapper(encrypted, func() string { return wrongKey })

	// Assert
	if err == nil {
		t.Error("Expected error when decrypting with wrong key, got nil")
	}

	// The error should be related to authentication failure in GCM
	if err != nil && !strings.Contains(err.Error(), "authentication") && !strings.Contains(err.Error(), "cipher") {
		// GCM will fail authentication, which may show as "message authentication failed" or similar
		t.Logf("Decryption with wrong key failed with error: %v", err)
	}
}

func TestEncryptString_MultilineText(t *testing.T) {
	// Arrange
	plainText := `This is a
multi-line
text with
several lines`

	// Act
	encrypted, err := EncryptStringWrapper(plainText, GetMockKey)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	decrypted, err := DecryptStringWrapper(encrypted, GetMockKey)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	// Assert
	if decrypted != plainText {
		t.Errorf("Expected multiline text to be preserved.\nExpected:\n%s\nGot:\n%s", plainText, decrypted)
	}
}

func TestEncryptString_JSONData(t *testing.T) {
	// Arrange
	jsonData := `{"name":"John","age":30,"city":"New York"}`

	// Act
	encrypted, err := EncryptStringWrapper(jsonData, GetMockKey)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	decrypted, err := DecryptStringWrapper(encrypted, GetMockKey)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	// Assert
	if decrypted != jsonData {
		t.Errorf("Expected JSON data to be preserved.\nExpected:\n%s\nGot:\n%s", jsonData, decrypted)
	}
}

func TestEncryptString_TestToken(t *testing.T) {
	// Arrange
	jsonData := "valid-token"

	// Act
	encrypted, err := EncryptStringWrapper(jsonData, GetMockKey)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	decrypted, err := DecryptStringWrapper(encrypted, GetMockKey)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	// Assert
	if decrypted != jsonData {
		t.Errorf("Expected JSON data to be preserved.\nExpected:\n%s\nGot:\n%s", jsonData, decrypted)
	}
}
