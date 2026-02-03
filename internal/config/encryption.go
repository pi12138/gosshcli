package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var keyFilePath string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home directory:", err)
		os.Exit(1)
	}
	configDir := filepath.Join(home, ".config", "gossh")
	keyFilePath = filepath.Join(configDir, "secret.key")
}

func getEncryptionKey() ([]byte, error) {
	if _, err := os.Stat(keyFilePath); os.IsNotExist(err) {
		key := make([]byte, 32) // AES-256
		if _, err := rand.Read(key); err != nil {
			return nil, fmt.Errorf("could not generate encryption key: %w", err)
		}
		if err := os.WriteFile(keyFilePath, key, 0600); err != nil {
			return nil, fmt.Errorf("could not save encryption key: %w", err)
		}
		return key, nil
	}
	return os.ReadFile(keyFilePath)
}

func Encrypt(plaintext string) (string, error) {
	key, err := getEncryptionKey()
	if err != nil {
		return "", err
	}

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

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return fmt.Sprintf("%x", ciphertext), nil
}

func Decrypt(ciphertextHex string) (string, error) {
	key, err := getEncryptionKey()
	if err != nil {
		return "", err
	}

	var ciphertext []byte
	fmt.Sscanf(ciphertextHex, "%x", &ciphertext)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
