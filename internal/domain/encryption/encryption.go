package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/domain/configuration"
)

func EncryptString(plainText string) (string, error) {
	return EncryptStringWrapper(plainText, configuration.GetEncryptionKey)
}

func EncryptStringNoErr(plainText string, getKey func() string) string {
	value, err := EncryptStringWrapper(plainText, getKey)
	if err != nil {
		return ""
	}
	return value
}

// EncryptString encrypts a plaintext string using AES-GCM symmetric encryption
func EncryptStringWrapper(plainText string, getKey func() string) (string, error) {
	key := getKey()

	// Create a new AES cipher block
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the plaintext
	ciphertext := gcm.Seal(nonce, nonce, []byte(plainText), nil)

	// Encode to unpadded URL-safe base64 for easy storage/transmission (avoids +, /, and = characters)
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

func DecryptString(encryptedText string) (string, error) {
	return DecryptStringWrapper(encryptedText, configuration.GetEncryptionKey)
}

func DecryptStringNoError(encryptedText string, getKey func() string) string {
	value, err := DecryptStringWrapper(encryptedText, getKey)
	if err != nil {
		return ""
	}
	return value
}

// DecryptString decrypts a base64-encoded encrypted string using AES-GCM symmetric encryption
func DecryptStringWrapper(encryptedText string, getKey func() string) (string, error) {
	key := getKey()

	// Decode from unpadded URL-safe base64
	ciphertext, err := base64.RawURLEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Verify ciphertext is long enough
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// Extract nonce and encrypted data
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
