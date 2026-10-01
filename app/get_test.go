package app_test

import (
	"testing"

	"github.com/egermano/balde/app"
	"github.com/egermano/balde/core"
)

func TestGetBucketReturnsDTO(t *testing.T) {
	store := newMemoryStore()
	mustCreateBucket(t, store, core.Bucket{Name: "goals", Target: 50000, Balance: 20000, BudgetID: "default"})
	a := app.New("default", store)

	bucket, err := a.GetBucket("bkt-1")
	if err != nil {
		t.Fatalf("GetBucket() error = %v, want nil", err)
	}
	want := app.Bucket{ID: "bkt-1", Name: "goals", Target: 50000, Balance: 20000, BudgetID: "default"}
	if bucket != want {
		t.Errorf("GetBucket() = %+v, want %+v", bucket, want)
	}
}

func TestGetBucketUnknownIDReturnsError(t *testing.T) {
	a := app.New("default", newMemoryStore())

	if _, err := a.GetBucket("nope"); err == nil {
		t.Error("GetBucket(unknown) = nil error, want error")
	}
}

func TestGetTransactionReturnsDTO(t *testing.T) {
	store := newMemoryStore()
	mustCreateAccount(t, store, core.Account{Name: "Main", Type: core.AccountChecking})
	date := date2026January15()
	mustCreateTransaction(t, store, core.Transaction{
		Amount: -2500, Description: "a", Date: date, AccountID: "acc-1",
	})
	a := app.New("default", store)

	txn, err := a.GetTransaction("txn-1")
	if err != nil {
		t.Fatalf("GetTransaction() error = %v, want nil", err)
	}
	want := app.Transaction{ID: "txn-1", Amount: -2500, Description: "a", Date: date, AccountID: "acc-1"}
	if txn != want {
		t.Errorf("GetTransaction() = %+v, want %+v", txn, want)
	}
}

func TestGetTransactionUnknownIDReturnsError(t *testing.T) {
	a := app.New("default", newMemoryStore())

	if _, err := a.GetTransaction("nope"); err == nil {
		t.Error("GetTransaction(unknown) = nil error, want error")
	}
}
