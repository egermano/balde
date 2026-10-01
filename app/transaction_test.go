package app_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/egermano/balde/app"
	"github.com/egermano/balde/core"
)

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
