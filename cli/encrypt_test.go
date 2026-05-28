package cli

import (
	"os"
	"testing"

	"github.com/egermano/balde/store"
)

func TestEncryptCmd_PlainDBToEncrypted(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create a plain DB
	cmd := newInitCmd()
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Verify it's plain
	config, _ := store.ReadConfig("balde.db")
	if config.Encrypted {
		t.Fatalf("initial DB should be plain")
	}

	// Encrypt it
	encryptCmd := newEncryptCmd()
	encryptCmd.Flags().Set("password", "newsecret")
	if err := encryptCmd.RunE(encryptCmd, []string{}); err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// Verify it's now encrypted
	config, _ = store.ReadConfig("balde.db")
	if !config.Encrypted {
		t.Fatalf("DB should be encrypted after encrypt command")
	}

	// Verify backup exists
	if _, err := os.Stat("balde.db.backup"); os.IsNotExist(err) {
		t.Fatalf("backup file should be created")
	}

	// Verify can open with new password
	s, err := store.OpenStore("balde.db", "newsecret", "", config)
	if err != nil {
		t.Fatalf("failed to open encrypted DB: %v", err)
	}
	defer s.Close()

	// Verify buckets are still there
	buckets, _ := s.ListBuckets()
	if len(buckets) != 6 {
		t.Fatalf("expected 6 buckets, got %d", len(buckets))
	}
}

func TestEncryptCmd_AlreadyEncrypted(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create encrypted DB
	cmd := newInitCmd()
	cmd.Flags().Set("password", "original")
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Try to encrypt again
	encryptCmd := newEncryptCmd()
	encryptCmd.Flags().Set("password", "newpassword")
	err := encryptCmd.RunE(encryptCmd, []string{})
	if err == nil || err.Error() != "database is already encrypted" {
		t.Fatalf("expected 'database is already encrypted' error, got: %v", err)
	}
}

func TestEncryptCmd_NoPassword(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create plain DB
	cmd := newInitCmd()
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Try to encrypt without password
	encryptCmd := newEncryptCmd()
	err := encryptCmd.RunE(encryptCmd, []string{})
	if err == nil {
		t.Fatalf("expected error for missing password")
	}
}

func TestEncryptCmd_BackupNameCollision(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create plain DB
	cmd := newInitCmd()
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Create a backup file
	os.WriteFile("balde.db.backup", []byte("old backup"), 0644)

	// Encrypt (should handle collision)
	encryptCmd := newEncryptCmd()
	encryptCmd.Flags().Set("password", "newsecret")
	if err := encryptCmd.RunE(encryptCmd, []string{}); err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// Verify DB is encrypted
	config, _ := store.ReadConfig("balde.db")
	if !config.Encrypted {
		t.Fatalf("DB should be encrypted")
	}
}