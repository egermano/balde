package store_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egermano/balde/core"
	"github.com/egermano/balde/store"
)

func TestSQLiteStore_EnforcesNormalizedBucketNameUniquenessPerBudget(t *testing.T) {
	s, cleanup := setupTestDB(t)
	defer cleanup()

	if err := s.CreateBucket(core.Bucket{Name: "Fixed Costs", BudgetID: "b1"}); err != nil {
		t.Fatalf("create bucket: %v", err)
	}
	if err := s.CreateBucket(core.Bucket{Name: "  fixed costs  ", BudgetID: "b1"}); err == nil {
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

func TestSQLiteStore_UpdateBucketEnforcesNormalizedNameUniqueness(t *testing.T) {
	s, cleanup := setupTestDB(t)
	defer cleanup()

	if err := s.CreateBucket(core.Bucket{Name: "Housing", BudgetID: "b1"}); err != nil {
		t.Fatalf("create housing: %v", err)
	}
	if err := s.CreateBucket(core.Bucket{Name: "Food", BudgetID: "b1"}); err != nil {
		t.Fatalf("create food: %v", err)
	}
	buckets, err := s.ListBuckets()
	if err != nil {
		t.Fatalf("list buckets: %v", err)
	}
	buckets[1].Name = " housing "
	if err := s.UpdateBucket(buckets[1]); err == nil {
		t.Fatal("expected duplicate bucket update error")
	}
}

func TestSQLiteStore_MigrationRejectsLegacyDuplicateBucketsWithoutDataLoss(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "legacy.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}
	_, err = db.Exec(`
		CREATE TABLE buckets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			target INTEGER NOT NULL DEFAULT 0,
			balance INTEGER NOT NULL DEFAULT 0,
			budget_id TEXT NOT NULL
		);
		INSERT INTO buckets (name, budget_id) VALUES ('Fixed Costs', 'b1'), (' fixed costs ', 'b1');
	`)
	if err != nil {
		t.Fatalf("seed legacy db: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close legacy db: %v", err)
	}

	if _, err := store.NewSQLiteStore(dbPath); err == nil {
		t.Fatal("expected migration to reject duplicate buckets")
	}

	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("reopen legacy db: %v", err)
	}
	defer db.Close()
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM buckets").Scan(&count); err != nil {
		t.Fatalf("count legacy buckets: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected both legacy buckets preserved, got %d", count)
	}
}

func setupTestDB(t *testing.T) (*store.SQLiteStore, func()) {
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

func TestSQLiteStore_CreateAndGetAccount(t *testing.T) {
	s, cleanup := setupTestDB(t)
	defer cleanup()

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
	if got.Type != core.AccountChecking {
		t.Errorf("expected type=checking, got %s", got.Type)
	}
	if got.Balance != 100000 {
		t.Errorf("expected balance=100000, got %d", got.Balance)
	}
	if got.ID == "" {
		t.Error("expected non-empty ID")
	}

	fetched, err := s.GetAccount(got.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if fetched.ID != got.ID {
		t.Errorf("expected ID=%s, got %s", got.ID, fetched.ID)
	}
}

func TestSQLiteStore_CreateAndGetBucket(t *testing.T) {
	s, cleanup := setupTestDB(t)
	defer cleanup()

	bk := core.Bucket{
		Name:     "housing",
		Target:   50000,
		Balance:  0,
		BudgetID: "b1",
	}
	if err := s.CreateBucket(bk); err != nil {
		t.Fatalf("create bucket: %v", err)
	}

	buckets, err := s.ListBuckets()
	if err != nil {
		t.Fatalf("list buckets: %v", err)
	}
	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(buckets))
	}

	got := buckets[0]
	if got.Name != "housing" {
		t.Errorf("expected name=housing, got %s", got.Name)
	}
	if got.Target != 50000 {
		t.Errorf("expected target=50000, got %d", got.Target)
	}
	if got.ID == "" {
		t.Error("expected non-empty ID")
	}

	fetched, err := s.GetBucket(got.ID)
	if err != nil {
		t.Fatalf("get bucket: %v", err)
	}
	if fetched.ID != got.ID {
		t.Errorf("expected ID=%s, got %s", got.ID, fetched.ID)
	}
}

func TestSQLiteStore_UpdateBucket(t *testing.T) {
	s, cleanup := setupTestDB(t)
	defer cleanup()

	s.CreateBucket(core.Bucket{
		Name:     "housing",
		Target:   50000,
		BudgetID: "b1",
	})

	buckets, _ := s.ListBuckets()
	bk := buckets[0]
	bk.Balance = 30000

	if err := s.UpdateBucket(bk); err != nil {
		t.Fatalf("update bucket: %v", err)
	}

	updated, err := s.GetBucket(bk.ID)
	if err != nil {
		t.Fatalf("get bucket: %v", err)
	}
	if updated.Balance != 30000 {
		t.Errorf("expected balance=30000, got %d", updated.Balance)
	}
}

func TestSQLiteStore_DeleteBucket(t *testing.T) {
	s, cleanup := setupTestDB(t)
	defer cleanup()

	s.CreateBucket(core.Bucket{Name: "housing", Target: 50000, BudgetID: "b1"})
	buckets, _ := s.ListBuckets()
	id := buckets[0].ID

	if err := s.DeleteBucket(id); err != nil {
		t.Fatalf("delete bucket: %v", err)
	}

	buckets, _ = s.ListBuckets()
	if len(buckets) != 0 {
		t.Errorf("expected 0 buckets after delete, got %d", len(buckets))
	}
}

func TestSQLiteStore_CreateAndGetTransaction(t *testing.T) {
	s, cleanup := setupTestDB(t)
	defer cleanup()

	tx := core.Transaction{
		Amount:      -50000,
		Description: "rent",
		Date:        parseDate("2025-01-15"),
		AccountID:   "acc-1",
		BucketID:    "bkt-1",
	}
	if err := s.CreateTransaction(tx); err != nil {
		t.Fatalf("create transaction: %v", err)
	}

	txs, err := s.ListTransactions()
	if err != nil {
		t.Fatalf("list transactions: %v", err)
	}
	if len(txs) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(txs))
	}

	got := txs[0]
	if got.Amount != -50000 {
		t.Errorf("expected amount=-50000, got %d", got.Amount)
	}
	if got.Description != "rent" {
		t.Errorf("expected description=rent, got %s", got.Description)
	}
	if got.AccountID != "acc-1" {
		t.Errorf("expected account_id=acc-1, got %s", got.AccountID)
	}
	if got.BucketID != "bkt-1" {
		t.Errorf("expected bucket_id=bkt-1, got %s", got.BucketID)
	}
	if got.ID == "" {
		t.Error("expected non-empty ID")
	}

	fetched, err := s.GetTransaction(got.ID)
	if err != nil {
		t.Fatalf("get transaction: %v", err)
	}
	if fetched.ID != got.ID {
		t.Errorf("expected ID=%s, got %s", got.ID, fetched.ID)
	}
}

func TestSQLiteStore_UpdateTransaction(t *testing.T) {
	s, cleanup := setupTestDB(t)
	defer cleanup()

	s.CreateTransaction(core.Transaction{
		Amount:      -50000,
		Description: "rent",
		Date:        parseDate("2025-01-15"),
		AccountID:   "acc-1",
		BucketID:    "bkt-1",
	})

	txs, _ := s.ListTransactions()
	tx := txs[0]
	tx.BucketID = "bkt-2"
	tx.Categorized = true

	if err := s.UpdateTransaction(tx); err != nil {
		t.Fatalf("update transaction: %v", err)
	}

	updated, err := s.GetTransaction(tx.ID)
	if err != nil {
		t.Fatalf("get transaction: %v", err)
	}
	if updated.BucketID != "bkt-2" {
		t.Errorf("expected bucket_id=bkt-2, got %s", updated.BucketID)
	}
	if !updated.Categorized {
		t.Error("expected categorized=true")
	}
}

func parseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func TestSQLiteStore_IntegrationWithBudget(t *testing.T) {
	s, cleanup := setupTestDB(t)
	defer cleanup()

	budget := core.NewBudget("b1", s)

	acc, err := budget.AddAccount("checking", core.AccountChecking, 200000)
	if err != nil {
		t.Fatalf("add account: %v", err)
	}

	bkt, err := budget.AddBucket("housing", 80000)
	if err != nil {
		t.Fatalf("add bucket: %v", err)
	}

	if err := budget.Allocate(bkt.ID, 80000); err != nil {
		t.Fatalf("allocate: %v", err)
	}

	rain, err := budget.Rain()
	if err != nil {
		t.Fatalf("rain: %v", err)
	}
	if rain != 120000 {
		t.Errorf("expected rain=120000, got %d", rain)
	}

	_ = acc
}

func TestSQLiteStore_UpdateAccount(t *testing.T) {
	s, cleanup := setupTestDB(t)
	defer cleanup()

	s.CreateAccount(core.Account{
		Name:    "checking",
		Type:    core.AccountChecking,
		Balance: 100000,
	})

	accounts, _ := s.ListAccounts()
	a := accounts[0]
	a.Balance = 150000

	if err := s.UpdateAccount(a); err != nil {
		t.Fatalf("update account: %v", err)
	}

	updated, err := s.GetAccount(a.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if updated.Balance != 150000 {
		t.Errorf("expected balance=150000, got %d", updated.Balance)
	}
}
