package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"go-starter/config"
	"fmt"
	"io"
)

var (
	ErrInvalidKey        = errors.New("invalid key: must be 32 bytes")
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	ErrDecryptionFailed  = errors.New("decryption failed")
)

type Secure struct {
	key       []byte
	gcm       cipher.AEAD
	nonceSize int
}

func NewSecure(cfg *config.SecureConfig) (*Secure, error) {

	key, err := base64.StdEncoding.DecodeString(cfg.Key)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &Secure{
		key:       key,
		gcm:       gcm,
		nonceSize: gcm.NonceSize(),
	}, nil
}

func (s *Secure) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", fmt.Errorf("plaintext cannot be empty")
	}

	nonce := make([]byte, s.nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %v", err)
	}

	ciphertext := s.gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *Secure) Decrypt(encodedCiphertext string) (string, error) {
	if encodedCiphertext == "" {
		return "", fmt.Errorf("encoded ciphertext cannot be empty")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %v", err)
	}

	minLength := s.nonceSize + 16
	if len(ciphertext) < minLength {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:s.nonceSize], ciphertext[s.nonceSize:]

	plaintext, err := s.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %v", err)
	}

	return string(plaintext), nil
}
