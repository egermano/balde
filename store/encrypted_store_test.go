package store_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/egermano/balde/core"
	"github.com/egermano/balde/store"
)

func TestEncryptedSQLiteStore_CreateAndGetAccount(t *testing.T) {
	dir := t.TempDir()
	encPath := filepath.Join(dir, "test.enc")
	password := "test-password"

	s, err := store.NewEncryptedSQLiteStore(encPath, password)
	if err != nil {
		t.Fatalf("open encrypted store: %v", err)
	}
	defer s.Close()

	a := core.Account{
		Name:    "checking",
		Type:    core.AccountChecking,
		Balance: 100000,
	}
	if err := s.CreateAccount(a); err != nil {
		t.Fatalf("create account: %v", err)
	}

	accounts, err := s.ListAccounts()
	if err != nil {
		t.Fatalf("list accounts: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(accounts))
	}

	got := accounts[0]
	if got.Name != "checking" {
		t.Errorf("expected name=checking, got %s", got.Name)
	}
	if got.Balance != 100000 {
		t.Errorf("expected balance=100000, got %d", got.Balance)
	}
}

func TestEncryptedSQLiteStore_PersistsEncrypted(t *testing.T) {
	dir := t.TempDir()
	encPath := filepath.Join(dir, "test.enc")
	password := "test-password"

	s1, err := store.NewEncryptedSQLiteStore(encPath, password)
	if err != nil {
		t.Fatalf("open encrypted store: %v", err)
	}

	s1.CreateAccount(core.Account{
		Name:    "checking",
		Type:    core.AccountChecking,
		Balance: 100000,
	})

	t.Logf("closing store 1")
	if err := s1.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}

	data, err := os.ReadFile(encPath)
	if err != nil {
		t.Fatalf("read encrypted file: %v", err)
	}

	t.Logf("encrypted file size: %d bytes", len(data))

	if len(data) == 0 {
		t.Fatal("encrypted file should not be empty")
	}

	if string(data[:10]) == "SQLite" {
		t.Fatal("file should be encrypted, not plain SQLite")
	}

	t.Logf("opening store 2")
	s2, err := store.NewEncryptedSQLiteStore(encPath, password)
	if err != nil {
		t.Fatalf("reopen encrypted store: %v", err)
	}
	defer s2.Close()

	t.Logf("listing accounts")
	accounts, err := s2.ListAccounts()
	if err != nil {
		t.Fatalf("list accounts: %v", err)
	}
	t.Logf("got %d accounts", len(accounts))
	if len(accounts) != 1 {
		t.Fatalf("expected 1 account after reopen, got %d", len(accounts))
	}
	t.Logf("account: %+v", accounts[0])
	if accounts[0].Balance != 100000 {
		t.Errorf("expected balance=100000 after reopen, got %d", accounts[0].Balance)
	}
}

func TestEncryptedSQLiteStore_WrongPassword(t *testing.T) {
	dir := t.TempDir()
	encPath := filepath.Join(dir, "test.enc")

	s1, err := store.NewEncryptedSQLiteStore(encPath, "password1")
	if err != nil {
		t.Fatalf("open encrypted store: %v", err)
	}

	s1.CreateAccount(core.Account{
		Name:    "checking",
		Type:    core.AccountChecking,
		Balance: 100000,
	})

	if err := s1.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}

	_, err = store.NewEncryptedSQLiteStore(encPath, "wrong-password")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}