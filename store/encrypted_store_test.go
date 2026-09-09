package store_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/egermano/balde/core"
	"github.com/egermano/balde/store"
)

func TestEncryptedSQLiteStore_EnforcesNormalizedBucketNameUniquenessPerBudget(t *testing.T) {
	encPath := filepath.Join(t.TempDir(), "test.enc")
	s, err := store.NewEncryptedSQLiteStore(encPath, "test-password")
	if err != nil {
		t.Fatalf("open encrypted store: %v", err)
	}
	defer s.Close()

	if err := s.CreateBucket(core.Bucket{Name: "Fixed Costs", BudgetID: "b1"}); err != nil {
		t.Fatalf("create bucket: %v", err)
	}
	if err := s.CreateBucket(core.Bucket{Name: " fixed costs ", BudgetID: "b1"}); err == nil {
		t.Fatal("expected duplicate bucket error")
	}
	if err := s.CreateBucket(core.Bucket{Name: "   ", BudgetID: "b1"}); err == nil {
		t.Fatal("expected empty bucket name error")
	}
	if err := s.CreateBucket(core.Bucket{Name: " fixed costs ", BudgetID: "b2"}); err != nil {
		t.Fatalf("same name in another budget: %v", err)
	}

	buckets, err := s.ListBuckets()
	if err != nil {
		t.Fatalf("list buckets: %v", err)
	}
	if buckets[1].Name != "fixed costs" {
		t.Fatalf("expected trimmed persisted name, got %q", buckets[1].Name)
	}
}

func TestEncryptedSQLiteStore_MigrationRejectsLegacyDuplicateBucketsWithoutDataLoss(t *testing.T) {
	encPath := filepath.Join(t.TempDir(), "legacy.enc")
	password := "test-password"
	dump := []byte(`
		CREATE TABLE buckets (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, target INTEGER NOT NULL DEFAULT 0, balance INTEGER NOT NULL DEFAULT 0, budget_id TEXT NOT NULL);
		INSERT INTO buckets VALUES (1, 'Fixed Costs', 0, 0, 'b1');
		INSERT INTO buckets VALUES (2, ' fixed costs ', 0, 0, 'b1');
	`)
	ciphertext, salt, nonce, err := store.Encrypt(dump, password)
	if err != nil {
		t.Fatalf("encrypt legacy dump: %v", err)
	}
	serialized, err := store.SerializeEncrypted(ciphertext, salt, nonce)
	if err != nil {
		t.Fatalf("serialize legacy dump: %v", err)
	}
	if err := os.WriteFile(encPath, serialized, 0600); err != nil {
		t.Fatalf("write legacy encrypted db: %v", err)
	}

	if _, err := store.NewEncryptedSQLiteStore(encPath, password); err == nil {
		t.Fatal("expected migration to reject duplicate buckets")
	}
	after, err := os.ReadFile(encPath)
	if err != nil {
		t.Fatalf("read legacy encrypted db: %v", err)
	}
	if !bytes.Equal(after, serialized) {
		t.Fatal("migration failure changed legacy encrypted data")
	}
}

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
