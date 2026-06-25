package cli

import (
	"os"
	"testing"

	"github.com/egermano/balde/store"
)

func TestInitCmd_NonInteractive_NoEncryptionFlag(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"init", "--no-encryption"})
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	if _, err := os.Stat("balde.json"); os.IsNotExist(err) {
		t.Fatalf("config file was not created")
	}

	if _, err := os.Stat("balde.db"); os.IsNotExist(err) {
		t.Fatalf("db file was not created")
	}

	// Should be plain (non-encrypted) database
	config, err := store.ReadConfig("balde.db")
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	if config.Encrypted {
		t.Fatalf("expected plain db with --no-encryption flag, got encrypted")
	}
}

func TestInitCmd_NonInteractive_EncryptionEnvVar(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Set environment variable to enable encryption
	os.Setenv("BALDE_ENCRYPTION", "true")
	defer os.Unsetenv("BALDE_ENCRYPTION")
	
	// Set password via environment variable
	os.Setenv("BALDE_PASSWORD", "test123")
	defer os.Unsetenv("BALDE_PASSWORD")

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"init", "--non-interactive"})
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	if _, err := os.Stat("balde.json"); os.IsNotExist(err) {
		t.Fatalf("config file was not created")
	}

	if _, err := os.Stat("balde.db"); os.IsNotExist(err) {
		t.Fatalf("encrypted db file was not created")
	}

	// Should be encrypted database
	config, err := store.ReadConfig("balde.db")
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	if !config.Encrypted {
		t.Fatalf("expected encrypted db with BALDE_ENCRYPTION=true, got plain")
	}

	// Verify we can open the encrypted store
	s, err := store.OpenStore("balde.db", "test123", "", config)
	if err != nil {
		t.Fatalf("failed to open encrypted store: %v", err)
	}
	defer s.Close()

	buckets, err := s.ListBuckets()
	if err != nil {
		t.Fatalf("failed to list buckets: %v", err)
	}

	if len(buckets) != 6 {
		t.Fatalf("expected 6 default buckets, got %d", len(buckets))
	}
}

func TestInitCmd_NonInteractive_NoEncryptionEnvVar(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Set environment variable to disable encryption
	os.Setenv("BALDE_ENCRYPTION", "false")
	defer os.Unsetenv("BALDE_ENCRYPTION")

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"init", "--non-interactive"})
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	if _, err := os.Stat("balde.json"); os.IsNotExist(err) {
		t.Fatalf("config file was not created")
	}

	if _, err := os.Stat("balde.db"); os.IsNotExist(err) {
		t.Fatalf("db file was not created")
	}

	// Should be plain (non-encrypted) database
	config, err := store.ReadConfig("balde.db")
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	if config.Encrypted {
		t.Fatalf("expected plain db with BALDE_ENCRYPTION=false, got encrypted")
	}
}

func TestInitCmd_NonInteractive_MissingPassword(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Set environment variable to enable encryption but no password
	os.Setenv("BALDE_ENCRYPTION", "true")
	defer os.Unsetenv("BALDE_ENCRYPTION")

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"init", "--non-interactive"})
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error when encryption enabled but no password provided")
	}

	expectedError := "password required when encryption enabled (use --password flag or BALDE_PASSWORD env var)"
	if err.Error() != expectedError {
		t.Fatalf("expected %q, got: %v", expectedError, err)
	}
}