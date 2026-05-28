package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/scrypt"
)

const (
	scryptN     = 32768
	scryptR     = 8
	scryptP     = 1
	keyLength   = 32
	saltLength  = 16
	nonceLength = 12
)

func DeriveKey(password string, salt []byte) ([]byte, error) {
	return scrypt.Key([]byte(password), salt, scryptN, scryptR, scryptP, keyLength)
}

func Encrypt(plaintext []byte, password string) (ciphertext, salt, nonce []byte, err error) {
	salt = make([]byte, saltLength)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, nil, nil, fmt.Errorf("generate salt: %w", err)
	}

	key, err := DeriveKey(password, salt)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("derive key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("aes cipher: %w", err)
	}

	nonce = make([]byte, nonceLength)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, nil, fmt.Errorf("generate nonce: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("gcm: %w", err)
	}

	ciphertext = aesgcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, salt, nonce, nil
}

func Decrypt(ciphertext, salt, nonce []byte, password string) ([]byte, error) {
	key, err := DeriveKey(password, salt)
	if err != nil {
		return nil, fmt.Errorf("derive key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	return plaintext, nil
}

func SerializeEncrypted(ciphertext, salt, nonce []byte) ([]byte, error) {
	if len(salt) != saltLength {
		return nil, fmt.Errorf("invalid salt length: %d", len(salt))
	}
	if len(nonce) != nonceLength {
		return nil, fmt.Errorf("invalid nonce length: %d", len(nonce))
	}

	result := make([]byte, saltLength+nonceLength+len(ciphertext))
	copy(result[:saltLength], salt)
	copy(result[saltLength:saltLength+nonceLength], nonce)
	copy(result[saltLength+nonceLength:], ciphertext)
	return result, nil
}

func DeserializeEncrypted(data []byte) (ciphertext, salt, nonce []byte, err error) {
	if len(data) < saltLength+nonceLength {
		return nil, nil, nil, fmt.Errorf("data too short: %d", len(data))
	}

	salt = data[:saltLength]
	nonce = data[saltLength : saltLength+nonceLength]
	ciphertext = data[saltLength+nonceLength:]
	return ciphertext, salt, nonce, nil
}
