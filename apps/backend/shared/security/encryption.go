package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
)

// deriveKey converts a secret key string into a strict 32-byte key for AES-256
func deriveKey(secret string) []byte {
	if len(secret) == 64 {
		if decoded, err := hex.DecodeString(secret); err == nil && len(decoded) == 32 {
			return decoded
		}
	}
	if len(secret) == 32 {
		return []byte(secret)
	}
	hash := sha256.Sum256([]byte(secret))
	return hash[:]
}

// EncryptAESGCM encrypts plain text using AES-GCM with a secret key loaded from env
func EncryptAESGCM(plainText string, secret string) (string, error) {
	if secret == "" {
		return "", errors.New("encryption key cannot be empty")
	}
	key := deriveKey(secret)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// DecryptAESGCM decrypts base64 encoded ciphertext using AES-GCM and the secret key
func DecryptAESGCM(cipherTextB64 string, secret string) (string, error) {
	if secret == "" {
		return "", errors.New("encryption key cannot be empty")
	}
	data, err := base64.StdEncoding.DecodeString(cipherTextB64)
	if err != nil {
		return "", err
	}

	key := deriveKey(secret)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, cipherText := data[:nonceSize], data[nonceSize:]
	plainBytes, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", errors.New("decryption failed: invalid key or corrupted data")
	}

	return string(plainBytes), nil
}
