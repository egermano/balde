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

func TestEncryptedSQLiteStore_CreateTransaction_AtomicallyAppliesBalances(t *testing.T) {
	dir := t.TempDir()
	s, err := store.NewEncryptedSQLiteStore(filepath.Join(dir, "test.enc"), "test-password")
	if err != nil {
		t.Fatalf("open encrypted store: %v", err)
	}
	defer s.Close()

	if err := s.CreateAccount(core.Account{Name: "checking", Type: core.AccountChecking, Balance: 10000}); err != nil {
		t.Fatalf("create account: %v", err)
	}
	if err := s.CreateBucket(core.Bucket{Name: "food", Target: 5000, BudgetID: "b1"}); err != nil {
		t.Fatalf("create bucket: %v", err)
	}
	if err := s.CreateTransaction(core.Transaction{
		Amount: -2500, Description: "groceries", AccountID: "1", BucketID: "1", Categorized: true,
	}); err != nil {
		t.Fatalf("create transaction: %v", err)
	}

	account, _ := s.GetAccount("1")
	bucket, _ := s.GetBucket("1")
	if account.Balance != 7500 {
		t.Errorf("expected account balance 7500, got %d", account.Balance)
	}
	if bucket.Balance != -2500 {
		t.Errorf("expected bucket balance -2500, got %d", bucket.Balance)
	}
}

func TestEncryptedSQLiteStore_DeleteTransaction_AtomicallyReversesBalances(t *testing.T) {
	dir := t.TempDir()
	s, err := store.NewEncryptedSQLiteStore(filepath.Join(dir, "test.enc"), "test-password")
	if err != nil {
		t.Fatalf("open encrypted store: %v", err)
	}
	defer s.Close()
	s.CreateAccount(core.Account{Name: "checking", Type: core.AccountChecking, Balance: 10000})
	s.CreateBucket(core.Bucket{Name: "food", Target: 5000, BudgetID: "b1"})
	if err := s.CreateTransaction(core.Transaction{
		Amount: -2500, Description: "groceries", AccountID: "1", BucketID: "1", Categorized: true,
	}); err != nil {
		t.Fatalf("create transaction: %v", err)
	}

	if err := s.DeleteTransaction("1"); err != nil {
		t.Fatalf("delete transaction: %v", err)
	}

	account, _ := s.GetAccount("1")
	bucket, _ := s.GetBucket("1")
	if account.Balance != 10000 {
		t.Errorf("expected account balance 10000, got %d", account.Balance)
	}
	if bucket.Balance != 0 {
		t.Errorf("expected bucket balance 0, got %d", bucket.Balance)
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
