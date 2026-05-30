package store_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/egermano/balde/store"
)

func setupEmptyTestDB(t *testing.T) (*store.SQLiteStore, func()) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	s, err := store.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	return s, func() {
		s.Close()
		os.RemoveAll(dir)
	}
}

func TestSQLiteStore_ListAccounts_Empty(t *testing.T) {
	s, cleanup := setupEmptyTestDB(t)
	defer cleanup()

	accounts, err := s.ListAccounts()
	if err != nil {
		t.Fatalf("list accounts: %v", err)
	}
	
	t.Logf("accounts result: %v (type: %T)", accounts, accounts)
	t.Logf("accounts == nil: %v", accounts == nil)
	t.Logf("len(accounts): %d", len(accounts))
	
	if accounts == nil {
		t.Error("expected empty slice, got nil")
	}
	
	if len(accounts) != 0 {
		t.Errorf("expected 0 accounts, got %d", len(accounts))
	}
}

func TestSQLiteStore_ListTransactions_Empty(t *testing.T) {
	s, cleanup := setupEmptyTestDB(t)
	defer cleanup()

	transactions, err := s.ListTransactions()
	if err != nil {
		t.Fatalf("list transactions: %v", err)
	}
	
	if transactions == nil {
		t.Error("expected empty slice, got nil")
	}
	
	if len(transactions) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(transactions))
	}
}

func TestSQLiteStore_ListBuckets_Empty(t *testing.T) {
	s, cleanup := setupEmptyTestDB(t)
	defer cleanup()

	buckets, err := s.ListBuckets()
	if err != nil {
		t.Fatalf("list buckets: %v", err)
	}
	
	if buckets == nil {
		t.Error("expected empty slice, got nil")
	}
	
	if len(buckets) != 0 {
		t.Errorf("expected 0 buckets, got %d", len(buckets))
	}
}