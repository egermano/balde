package cli

import (
	"os"
	"testing"

	"github.com/egermano/balde/store"
)

func TestUnlockCmd_PlainDB(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create a plain DB
	cmd := newInitCmd()
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Try to unlock plain DB (should fail or do nothing)
	unlockCmd := newUnlockCmd()
	err := unlockCmd.RunE(unlockCmd, []string{})
	if err == nil || err.Error() != "database is not encrypted" {
		t.Fatalf("expected 'database is not encrypted' error, got: %v", err)
	}
}

func TestUnlockCmd_EncryptedDB_CorrectPassword(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create encrypted DB
	initCmd := newInitCmd()
	initCmd.Flags().Set("password", "mysecret")
	if err := initCmd.RunE(initCmd, []string{}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Remove session to simulate fresh unlock
	sessionPath := store.GetSessionPath("balde.db")
	os.Remove(sessionPath)

	// Unlock with correct password
	unlockCmd := newUnlockCmd()
	unlockCmd.Flags().Set("password", "mysecret")
	if err := unlockCmd.RunE(unlockCmd, []string{}); err != nil {
		t.Fatalf("unlock failed: %v", err)
	}

	// Verify session was created
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		t.Fatalf("session file was not created")
	}
}

func TestUnlockCmd_EncryptedDB_WrongPassword(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create encrypted DB
	initCmd := newInitCmd()
	initCmd.Flags().Set("password", "mysecret")
	if err := initCmd.RunE(initCmd, []string{}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Remove session to simulate fresh unlock
	sessionPath := store.GetSessionPath("balde.db")
	os.Remove(sessionPath)

	// Try to unlock with wrong password
	unlockCmd := newUnlockCmd()
	unlockCmd.Flags().Set("password", "wrongpassword")
	err := unlockCmd.RunE(unlockCmd, []string{})
	if err == nil {
		t.Fatalf("expected error for wrong password, got nil")
	}

	// Verify session was NOT created
	if _, err := os.Stat(sessionPath); !os.IsNotExist(err) {
		t.Fatalf("session file should not exist after wrong password")
	}
}

func TestUnlockCmd_EnvPassword(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create encrypted DB
	initCmd := newInitCmd()
	initCmd.Flags().Set("password", "mysecret")
	if err := initCmd.RunE(initCmd, []string{}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Remove session
	sessionPath := store.GetSessionPath("balde.db")
	os.Remove(sessionPath)

	// Set env var
	os.Setenv("BALDE_PASSWORD", "mysecret")
	defer os.Unsetenv("BALDE_PASSWORD")

	// Unlock with env password
	unlockCmd := newUnlockCmd()
	if err := unlockCmd.RunE(unlockCmd, []string{}); err != nil {
		t.Fatalf("unlock with env password failed: %v", err)
	}

	// Verify session was created
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		t.Fatalf("session file was not created")
	}
}

func TestUnlockCmd_NoPassword(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create encrypted DB
	initCmd := newInitCmd()
	initCmd.Flags().Set("password", "mysecret")
	if err := initCmd.RunE(initCmd, []string{}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Remove session
	sessionPath := store.GetSessionPath("balde.db")
	os.Remove(sessionPath)

	// Try to unlock without password
	unlockCmd := newUnlockCmd()
	err := unlockCmd.RunE(unlockCmd, []string{})
	if err == nil {
		t.Fatalf("expected error for no password, got nil")
	}

	// Verify session was not created
	if _, err := os.Stat(sessionPath); !os.IsNotExist(err) {
		t.Fatalf("session file should not exist after failed unlock")
	}
}
