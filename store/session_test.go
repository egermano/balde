package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSession_Create(t *testing.T) {
	tmpDir := t.TempDir()

	session := Session{
		Password:      "my-password-123",
		LastAccessed: time.Now(),
	}

	sessionPath := filepath.Join(tmpDir, "balde-session-12345")

	err := WriteSession(sessionPath, session, 30*time.Minute)
	if err != nil {
		t.Fatalf("failed to write session: %v", err)
	}

	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		t.Fatalf("session file was not created")
	}

	var loaded Session
	loaded, err = ReadSession(sessionPath)
	if err != nil {
		t.Fatalf("failed to read session: %v", err)
	}

	if len(loaded.Password) != len(session.Password) {
		t.Fatalf("password length mismatch: got %d, want %d", len(loaded.Password), len(session.Password))
	}

	if loaded.Password != session.Password {
		t.Fatalf("password mismatch: got %s, want %s", loaded.Password, session.Password)
	}
}

func TestSession_ReadExpired(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "balde-session-12345")

	oldTime := time.Now().Add(-31 * time.Minute)
	session := Session{
		Password:      "my-password-123",
		LastAccessed: oldTime, // 31 minutes ago
	}

	sf := sessionFile{
		Password:      session.Password,
		LastAccessed: oldTime.Unix(),
	}

	data, err := json.Marshal(sf)
	if err != nil {
		t.Fatalf("failed to marshal session: %v", err)
	}

	err = os.WriteFile(sessionPath, data, 0600)
	if err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	_, err = ReadSession(sessionPath)
	if err == nil {
		t.Fatalf("expected error for expired session, got nil")
	}

	if err != ErrSessionExpired {
		t.Fatalf("expected ErrSessionExpired, got: %v", err)
	}
}

func TestSession_Renew(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "balde-session-12345")

	session := Session{
		Password:      "my-password-123",
		LastAccessed: time.Now().Add(-29 * time.Minute), // 29 minutes ago, still valid
	}

	err := WriteSession(sessionPath, session, 30*time.Minute)
	if err != nil {
		t.Fatalf("failed to write session: %v", err)
	}

	err = RenewSession(sessionPath, 30*time.Minute)
	if err != nil {
		t.Fatalf("failed to renew session: %v", err)
	}

	loaded, err := ReadSession(sessionPath)
	if err != nil {
		t.Fatalf("failed to read renewed session: %v", err)
	}

	timeDiff := time.Since(loaded.LastAccessed)
	if timeDiff > 5*time.Second {
		t.Fatalf("last accessed time not updated: %v ago", timeDiff)
	}
}

func TestSession_GetSessionPath(t *testing.T) {
	dbPath := "/path/to/budget.db"

	sessionPath := GetSessionPath(dbPath)

	expected := filepath.Join(os.TempDir(), "balde-session-"+hashPath(dbPath))
	if sessionPath != expected {
		t.Fatalf("session path mismatch: got %s, want %s", sessionPath, expected)
	}
}

func TestSession_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "balde-session-12345")

	err := os.WriteFile(sessionPath, []byte("invalid json"), 0600)
	if err != nil {
		t.Fatalf("failed to write invalid file: %v", err)
	}

	_, err = ReadSession(sessionPath)
	if err == nil {
		t.Fatalf("expected error for invalid session file, got nil")
	}
}