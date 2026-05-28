package cli

import (
	"os"
	"testing"

	"github.com/egermano/balde/store"
)

func TestLockCmd_PlainDB(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create a plain DB
	cmd := newInitCmd()
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Try to lock plain DB (should fail)
	lockCmd := newLockCmd()
	err := lockCmd.RunE(lockCmd, []string{})
	if err == nil || err.Error() != "database is not encrypted" {
		t.Fatalf("expected 'database is not encrypted' error, got: %v", err)
	}
}

func TestLockCmd_EncryptedDB_NoSession(t *testing.T) {
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

	// Try to lock without session (should fail)
	lockCmd := newLockCmd()
	err := lockCmd.RunE(lockCmd, []string{})
	if err == nil || err.Error() != "no active session" {
		t.Fatalf("expected 'no active session' error, got: %v", err)
	}
}

func TestLockCmd_EncryptedDB_WithSession(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create encrypted DB with session
	initCmd := newInitCmd()
	initCmd.Flags().Set("password", "mysecret")
	if err := initCmd.RunE(initCmd, []string{}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	sessionPath := store.GetSessionPath("balde.db")
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		t.Fatalf("session file should exist after init")
	}

	// Lock the session
	lockCmd := newLockCmd()
	if err := lockCmd.RunE(lockCmd, []string{}); err != nil {
		t.Fatalf("lock failed: %v", err)
	}

	// Verify session was deleted
	if _, err := os.Stat(sessionPath); !os.IsNotExist(err) {
		t.Fatalf("session file should be deleted after lock")
	}
}

func TestLockCmd_AfterLock_CannotAccessWithoutPassword(t *testing.T) {
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

	// Lock
	lockCmd := newLockCmd()
	if err := lockCmd.RunE(lockCmd, []string{}); err != nil {
		t.Fatalf("lock failed: %v", err)
	}

	// Try to open without password (should fail)
	config, _ := store.ReadConfig("balde.db")
	_, err := store.OpenStore("balde.db", "", "", config)
	if err == nil {
		t.Fatalf("should not be able to open encrypted DB without password after lock")
	}
}
