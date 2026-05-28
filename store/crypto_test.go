package store_test

import (
	"testing"

	"github.com/egermano/balde/store"
)

func TestEncryptDecrypt_Roundtrip(t *testing.T) {
	password := "my-secret-password"
	plaintext := []byte("This is sensitive financial data")

	ciphertext, salt, nonce, err := store.Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	decrypted, err := store.Decrypt(ciphertext, salt, nonce, password)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("roundtrip failed: got %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptDecrypt_WrongPassword(t *testing.T) {
	password := "my-secret-password"
	plaintext := []byte("This is sensitive financial data")

	ciphertext, salt, nonce, err := store.Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	_, err = store.Decrypt(ciphertext, salt, nonce, "wrong-password")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestEncryptDecrypt_TamperedCiphertext(t *testing.T) {
	password := "my-secret-password"
	plaintext := []byte("This is sensitive financial data")

	ciphertext, salt, nonce, err := store.Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	ciphertext[0] ^= 0xFF // flip a bit

	_, err = store.Decrypt(ciphertext, salt, nonce, password)
	if err == nil {
		t.Fatal("expected error for tampered ciphertext")
	}
}

func TestDeriveKey_SamePasswordSameKey(t *testing.T) {
	password := "my-password"
	salt1 := make([]byte, 16)
	salt2 := make([]byte, 16)

	key1, err := store.DeriveKey(password, salt1)
	if err != nil {
		t.Fatalf("derive key 1 failed: %v", err)
	}

	key2, err := store.DeriveKey(password, salt2)
	if err != nil {
		t.Fatalf("derive key 2 failed: %v", err)
	}

	if len(key1) != 32 {
		t.Errorf("expected 32-byte key, got %d", len(key1))
	}
	if len(key2) != 32 {
		t.Errorf("expected 32-byte key, got %d", len(key2))
	}
}

func TestDeriveKey_DifferentPasswordDifferentKey(t *testing.T) {
	salt := make([]byte, 16)

	key1, err := store.DeriveKey("password1", salt)
	if err != nil {
		t.Fatalf("derive key 1 failed: %v", err)
	}

	key2, err := store.DeriveKey("password2", salt)
	if err != nil {
		t.Fatalf("derive key 2 failed: %v", err)
	}

	if string(key1) == string(key2) {
		t.Error("different passwords should derive different keys")
	}
}

func TestDeriveKey_DifferentSaltDifferentKey(t *testing.T) {
	password := "my-password"
	salt1 := make([]byte, 16)
	salt2 := make([]byte, 16)
	salt2[0] = 1

	key1, err := store.DeriveKey(password, salt1)
	if err != nil {
		t.Fatalf("derive key 1 failed: %v", err)
	}

	key2, err := store.DeriveKey(password, salt2)
	if err != nil {
		t.Fatalf("derive key 2 failed: %v", err)
	}

	if string(key1) == string(key2) {
		t.Error("different salts should derive different keys")
	}
}
