package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flex-service/config"
	"fmt"
)

var (
	ErrInvalidKey        = errors.New("encryption key is empty or not set")
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	ErrDecryptionFailed  = errors.New("decryption failed")
	ErrHMACMismatch      = errors.New("HMAC verification failed")
)

type Secure struct {
	key []byte
}

// NewSecure creates a new Secure instance using the ENCRYPT_KEY environment variable
func NewSecure(cfg *config.SecureConfig) (*Secure, error) {
	key := cfg.Key
	if key == "" {
		return nil, ErrInvalidKey
	}

	return &Secure{
		key: []byte(key),
	}, nil
}

// Encrypt encrypts a string using AES-256-CBC with HMAC authentication
// Compatible with the PHP encrypt function
func (s *Secure) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", fmt.Errorf("plaintext cannot be empty")
	}

	// Create AES cipher
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	// Generate random IV
	iv := make([]byte, aes.BlockSize) // 16 bytes for AES
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %v", err)
	}

	// Pad plaintext to block size
	paddedPlaintext := pkcs7Pad([]byte(plaintext), aes.BlockSize)

	// Encrypt using CBC mode
	mode := cipher.NewCBCEncrypter(block, iv)
	encrypted := make([]byte, len(paddedPlaintext))
	mode.CryptBlocks(encrypted, paddedPlaintext)

	// Generate HMAC
	h := hmac.New(sha256.New, s.key)
	h.Write(encrypted)
	hmacSum := h.Sum(nil)

	// Combine: IV + HMAC + encrypted data
	result := make([]byte, 0, len(iv)+len(hmacSum)+len(encrypted))
	result = append(result, iv...)
	result = append(result, hmacSum...)
	result = append(result, encrypted...)

	// Return as hex string (like PHP bin2hex)
	return hex.EncodeToString(result), nil
}

// Decrypt decrypts a hex-encoded string using AES-256-CBC with HMAC verification
// Compatible with the PHP decrypt function
func (s *Secure) Decrypt(hexCiphertext string) (string, error) {
	if hexCiphertext == "" {
		return "", fmt.Errorf("ciphertext cannot be empty")
	}

	// Decode hex string (like PHP hex2bin)
	ciphertext, err := hex.DecodeString(hexCiphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex: %v", err)
	}

	// Check minimum length: IV(16) + HMAC(32) + at least one block(16) = 64 bytes
	if len(ciphertext) < 64 {
		return "", ErrInvalidCiphertext
	}

	// Extract components
	iv := ciphertext[:16]             // First 16 bytes
	receivedHMAC := ciphertext[16:48] // Next 32 bytes (SHA256 hash)
	encrypted := ciphertext[48:]      // Remaining bytes

	// Verify HMAC
	h := hmac.New(sha256.New, s.key)
	h.Write(encrypted)
	calculatedHMAC := h.Sum(nil)

	if !hmac.Equal(receivedHMAC, calculatedHMAC) {
		return "", ErrHMACMismatch
	}

	// Create AES cipher
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	// Decrypt using CBC mode
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(encrypted))
	mode.CryptBlocks(decrypted, encrypted)

	// Remove PKCS7 padding
	unpaddedPlaintext, err := pkcs7Unpad(decrypted, aes.BlockSize)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(unpaddedPlaintext), nil
}

// pkcs7Pad adds PKCS7 padding to data
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := make([]byte, padding)
	for i := range padText {
		padText[i] = byte(padding)
	}
	return append(data, padText...)
}

// pkcs7Unpad removes PKCS7 padding from data
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("invalid padding")
	}

	padding := int(data[len(data)-1])
	if padding == 0 || padding > blockSize {
		return nil, errors.New("invalid padding")
	}

	if len(data) < padding {
		return nil, errors.New("invalid padding")
	}

	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, errors.New("invalid padding")
		}
	}

	return data[:len(data)-padding], nil
}
