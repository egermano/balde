package cli

import (
	"os"
	"testing"

	"github.com/egermano/balde/store"
)

func TestInitCmd_PromptPassword(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	cmd := newInitCmd()
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	if _, err := os.Stat("balde.json"); os.IsNotExist(err) {
		t.Fatalf("config file was not created")
	}

	if _, err := os.Stat("balde.db"); os.IsNotExist(err) {
		t.Fatalf("db file was not created")
	}

	config, err := store.ReadConfig("balde.db")
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	if config.Encrypted {
		t.Fatalf("expected plain db, got encrypted")
	}
}

func TestInitCmd_WithPassword(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	cmd := newInitCmd()
	cmd.Flags().Set("password", "test123")
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	if _, err := os.Stat("balde.json"); os.IsNotExist(err) {
		t.Fatalf("config file was not created")
	}

	if _, err := os.Stat("balde.db"); os.IsNotExist(err) {
		t.Fatalf("encrypted db file was not created")
	}

	config, err := store.ReadConfig("balde.db")
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	if !config.Encrypted {
		t.Fatalf("expected encrypted db, got plain")
	}

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

func TestInitCmd_AlreadyInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	cmd := newInitCmd()
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("first init failed: %v", err)
	}

	cmd2 := newInitCmd()
	err := cmd2.RunE(cmd2, []string{})
	if err == nil {
		t.Fatalf("expected error for reinitialization, got nil")
	}

	if err.Error() != "budget already initialized in this directory" {
		t.Fatalf("expected 'budget already initialized in this directory' error, got: %v", err)
	}
}

func TestInitCmd_DifferentDir(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	cmd := newInitCmd()
	cmd.Flags().Set("dir", "budget")
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	if _, err := os.Stat("budget/balde.json"); os.IsNotExist(err) {
		t.Fatalf("config file was not created in budget dir")
	}

	if _, err := os.Stat("budget/balde.db"); os.IsNotExist(err) {
		t.Fatalf("db file was not created in budget dir")
	}

	files, _ := os.ReadDir(tmpDir)
	for _, f := range files {
		t.Logf("File in tmpdir: %s", f.Name())
	}

	budgetFiles, _ := os.ReadDir("budget")
	for _, f := range budgetFiles {
		t.Logf("File in budget dir: %s", f.Name())
	}
}