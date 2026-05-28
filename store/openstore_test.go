package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/egermano/balde/core"
)

func TestOpenStore_PlainDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "budget.db")

	config := Config{
		Encrypted: false,
	}

	store, err := OpenStore(dbPath, "", "", config)
	if err != nil {
		t.Fatalf("failed to open plain store: %v", err)
	}
	defer store.Close()

	if _, ok := store.(*SQLiteStore); !ok {
		t.Fatalf("expected SQLiteStore, got %T", store)
	}
}

func TestOpenStore_EncryptedDB_NoPassword(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "budget.db")

	config := Config{
		Encrypted: true,
	}

	_, err := OpenStore(dbPath, "", "", config)
	if err == nil {
		t.Fatalf("expected error for encrypted db without password")
	}
}

func TestOpenStore_EncryptedDB_NewDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "budget.db")

	config := Config{
		Encrypted: true,
	}

	store, err := OpenStore(dbPath, "password123", "", config)
	if err != nil {
		t.Fatalf("failed to open encrypted store: %v", err)
	}

	account := core.Account{Name: "Test Account", Type: "checking", Balance: 1000}
	if err := store.CreateAccount(account); err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("failed to close store: %v", err)
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatalf("encrypted file was not created")
	}
}

func TestOpenStore_EncryptedDB_ExistingDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "budget.db")

	config := Config{
		Encrypted: true,
	}

	store1, err := OpenStore(dbPath, "password123", "", config)
	if err != nil {
		t.Fatalf("failed to create encrypted store: %v", err)
	}

	account := core.Account{Name: "Test Account", Type: "checking", Balance: 1000}
	if err := store1.CreateAccount(account); err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	if err := store1.Close(); err != nil {
		t.Fatalf("failed to close store1: %v", err)
	}

	store2, err := OpenStore(dbPath, "password123", "", config)
	if err != nil {
		t.Fatalf("failed to reopen encrypted store: %v", err)
	}
	defer store2.Close()

	accounts, err := store2.ListAccounts()
	if err != nil {
		t.Fatalf("failed to list accounts: %v", err)
	}

	if len(accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(accounts))
	}

	if accounts[0].Name != "Test Account" {
		t.Fatalf("expected account name 'Test Account', got '%s'", accounts[0].Name)
	}
}

func TestOpenStore_EncryptedDB_WrongPassword(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "budget.db")

	config := Config{
		Encrypted: true,
	}

	store1, err := OpenStore(dbPath, "password123", "", config)
	if err != nil {
		t.Fatalf("failed to create encrypted store: %v", err)
	}

	if err := store1.Close(); err != nil {
		t.Fatalf("failed to close store1: %v", err)
	}

	_, err = OpenStore(dbPath, "wrongpassword", "", config)
	if err == nil {
		t.Fatalf("expected error for wrong password")
	}
}

func TestOpenStore_SessionFile(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "budget.db")

	config := Config{
		Encrypted: true,
	}

	store1, err := OpenStore(dbPath, "password123", "", config)
	if err != nil {
		t.Fatalf("failed to create encrypted store: %v", err)
	}

	account := core.Account{Name: "Test Account", Type: "checking", Balance: 1000}
	if err := store1.CreateAccount(account); err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	sessionPath := GetSessionPath(dbPath)
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		t.Fatalf("session file was not created")
	}

	if err := store1.Close(); err != nil {
		t.Fatalf("failed to close store1: %v", err)
	}

	store2, err := OpenStore(dbPath, "", "", config)
	if err != nil {
		t.Fatalf("failed to open store using session: %v", err)
	}

	accounts, err := store2.ListAccounts()
	if err != nil {
		t.Fatalf("failed to list accounts: %v", err)
	}

	if len(accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(accounts))
	}

	if accounts[0].Name != "Test Account" {
		t.Fatalf("expected account name 'Test Account', got '%s'", accounts[0].Name)
	}

	if err := store2.Close(); err != nil {
		t.Fatalf("failed to close store2: %v", err)
	}
}