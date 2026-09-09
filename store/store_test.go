package store_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egermano/balde/core"
	"github.com/egermano/balde/store"
)

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
	if err := s.CreateAccount(core.Account{Name: "checking", Type: core.AccountChecking}); err != nil {
		t.Fatalf("create account: %v", err)
	}
	if err := s.CreateBucket(core.Bucket{Name: "housing", BudgetID: "b1"}); err != nil {
		t.Fatalf("create bucket: %v", err)
	}

	tx := core.Transaction{
		Amount:      -50000,
		Description: "rent",
		Date:        parseDate("2025-01-15"),
		AccountID:   "1",
		BucketID:    "1",
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
	if got.AccountID != "1" {
		t.Errorf("expected account_id=1, got %s", got.AccountID)
	}
	if got.BucketID != "1" {
		t.Errorf("expected bucket_id=1, got %s", got.BucketID)
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

func TestSQLiteStore_CreateTransaction_AtomicallyAppliesBalances(t *testing.T) {
	s, cleanup := setupTestDB(t)
	defer cleanup()

	if err := s.CreateAccount(core.Account{Name: "checking", Type: core.AccountChecking, Balance: 10000}); err != nil {
		t.Fatalf("create account: %v", err)
	}
	if err := s.CreateBucket(core.Bucket{Name: "food", Target: 5000, BudgetID: "b1"}); err != nil {
		t.Fatalf("create bucket: %v", err)
	}

	if err := s.CreateTransaction(core.Transaction{
		Amount: -2500, Description: "groceries", Date: parseDate("2025-01-15"),
		AccountID: "1", BucketID: "1", Categorized: true,
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

func TestSQLiteStore_CreateTransaction_RollsBackWhenBucketDoesNotExist(t *testing.T) {
	s, cleanup := setupTestDB(t)
	defer cleanup()

	if err := s.CreateAccount(core.Account{Name: "checking", Type: core.AccountChecking, Balance: 10000}); err != nil {
		t.Fatalf("create account: %v", err)
	}

	err := s.CreateTransaction(core.Transaction{
		Amount: -2500, Description: "groceries", Date: parseDate("2025-01-15"),
		AccountID: "1", BucketID: "missing", Categorized: true,
	})
	if err == nil {
		t.Fatal("expected create transaction to fail")
	}

	account, _ := s.GetAccount("1")
	if account.Balance != 10000 {
		t.Errorf("expected unchanged account balance 10000, got %d", account.Balance)
	}
	transactions, _ := s.ListTransactions()
	if len(transactions) != 0 {
		t.Errorf("expected no transaction, got %d", len(transactions))
	}
}

func TestSQLiteStore_DeleteTransaction_AtomicallyReversesBalances(t *testing.T) {
	s, cleanup := setupTestDB(t)
	defer cleanup()
	s.CreateAccount(core.Account{Name: "checking", Type: core.AccountChecking, Balance: 10000})
	s.CreateBucket(core.Bucket{Name: "food", Target: 5000, BudgetID: "b1"})
	if err := s.CreateTransaction(core.Transaction{
		Amount: -2500, Description: "groceries", Date: parseDate("2025-01-15"),
		AccountID: "1", BucketID: "1", Categorized: true,
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
	if _, err := s.GetTransaction("1"); err == nil {
		t.Error("expected transaction to be deleted")
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
