package app_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/egermano/balde/app"
	"github.com/egermano/balde/core"
)

func TestAddTransactionCategorizedUpdatesBalances(t *testing.T) {
	store := newMemoryStore()
	mustCreateAccount(t, store, core.Account{Name: "Main", Type: core.AccountChecking, Balance: 100000})
	mustCreateBucket(t, store, core.Bucket{Name: "comfort", Target: 50000, Balance: 20000, BudgetID: "default"})
	a := app.New("default", store)

	date := time.Date(2026, 1, 16, 8, 0, 0, 0, time.UTC)
	txn, err := a.AddTransaction(-5000, "coffee beans", date, "acc-1", "bkt-1")
	if err != nil {
		t.Fatalf("AddTransaction() error = %v, want nil", err)
	}

	want := app.Transaction{
		ID:          "txn-1",
		Amount:      -5000,
		Description: "coffee beans",
		Date:        date,
		AccountID:   "acc-1",
		BucketID:    "bkt-1",
		Categorized: true,
	}
	if txn != want {
		t.Errorf("AddTransaction() = %+v, want %+v", txn, want)
	}

	status, err := a.Status()
	if err != nil {
		t.Fatalf("Status() error = %v, want nil", err)
	}
	if got := status.Accounts[0].Balance; got != 95000 {
		t.Errorf("account balance = %d, want 95000", got)
	}
	if got := status.Buckets[0].Balance; got != 15000 {
		t.Errorf("bucket balance = %d, want 15000", got)
	}
	if got := status.Rain; got != 80000 {
		t.Errorf("rain = %d, want 80000 (categorized spend must not change rain)", got)
	}
}

func TestAddTransactionUncategorizedIncomeRaisesRain(t *testing.T) {
	store := newMemoryStore()
	mustCreateAccount(t, store, core.Account{Name: "Main", Type: core.AccountChecking, Balance: 100000})
	mustCreateBucket(t, store, core.Bucket{Name: "goals", Target: 50000, Balance: 20000, BudgetID: "default"})
	a := app.New("default", store)

	txn, err := a.AddTransaction(50000, "tax refund", time.Date(2026, 2, 1, 9, 0, 0, 0, time.UTC), "acc-1", "")
	if err != nil {
		t.Fatalf("AddTransaction() error = %v, want nil", err)
	}
	if txn.Categorized {
		t.Error("AddTransaction() Categorized = true, want false for empty bucket")
	}

	status, err := a.Status()
	if err != nil {
		t.Fatalf("Status() error = %v, want nil", err)
	}
	if got := status.Rain; got != 130000 {
		t.Errorf("rain = %d, want 130000", got)
	}
}

func TestListTransactionsReturnsDTOs(t *testing.T) {
	store := newMemoryStore()
	mustCreateAccount(t, store, core.Account{Name: "Main", Type: core.AccountChecking})
	mustCreateBucket(t, store, core.Bucket{Name: "comfort", Target: 10000, BudgetID: "default"})
	date := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	mustCreateTransaction(t, store, core.Transaction{
		Amount:      -4200,
		Description: "groceries",
		Date:        date,
		AccountID:   "acc-1",
		BucketID:    "bkt-1",
		Categorized: true,
	})
	a := app.New("default", store)

	transactions, err := a.ListTransactions()
	if err != nil {
		t.Fatalf("ListTransactions() error = %v, want nil", err)
	}

	if len(transactions) != 1 {
		t.Fatalf("ListTransactions() returned %d transactions, want 1", len(transactions))
	}
	want := app.Transaction{
		ID:          "txn-1",
		Amount:      -4200,
		Description: "groceries",
		Date:        date,
		AccountID:   "acc-1",
		BucketID:    "bkt-1",
		Categorized: true,
	}
	if transactions[0] != want {
		t.Errorf("ListTransactions()[0] = %+v, want %+v", transactions[0], want)
	}
}

func TestDeleteTransactionRestoresBalances(t *testing.T) {
	store := newMemoryStore()
	mustCreateAccount(t, store, core.Account{Name: "Main", Type: core.AccountChecking, Balance: 100000})
	mustCreateBucket(t, store, core.Bucket{Name: "comfort", Target: 10000, Balance: 20000, BudgetID: "default"})
	mustCreateTransaction(t, store, core.Transaction{
		Amount: -2500, Description: "a", Date: time.Now(),
		AccountID: "acc-1", BucketID: "bkt-1", Categorized: true,
	})
	a := app.New("default", store)

	deleted, err := a.DeleteTransaction("txn-1")
	if err != nil {
		t.Fatalf("DeleteTransaction() error = %v, want nil", err)
	}
	if deleted.Amount != -2500 || deleted.Description != "a" {
		t.Errorf("DeleteTransaction() = %+v, want the deleted txn", deleted)
	}

	status, err := a.Status()
	if err != nil {
		t.Fatalf("Status() error = %v, want nil", err)
	}
	if got := status.Accounts[0].Balance; got != 100000 {
		t.Errorf("account balance = %d, want 100000", got)
	}
	if got := status.Buckets[0].Balance; got != 20000 {
		t.Errorf("bucket balance = %d, want 20000", got)
	}
	if len(status.Transactions) != 0 {
		t.Errorf("transactions after delete = %d, want 0", len(status.Transactions))
	}
}

func TestDeleteTransactionUnknownIDReturnsError(t *testing.T) {
	a := app.New("default", newMemoryStore())

	if _, err := a.DeleteTransaction("nope"); err == nil {
		t.Error("DeleteTransaction(unknown) = nil error, want error")
	}
}

func TestListTransactionsEmptyBudgetSerializesAsArray(t *testing.T) {
	a := app.New("default", newMemoryStore())

	transactions, err := a.ListTransactions()
	if err != nil {
		t.Fatalf("ListTransactions() error = %v, want nil", err)
	}

	data, err := json.Marshal(transactions)
	if err != nil {
		t.Fatalf("marshal transactions: %v", err)
	}
	if string(data) != "[]" {
		t.Errorf("ListTransactions() JSON = %s, want []", data)
	}
}
