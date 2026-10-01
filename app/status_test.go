package app_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/egermano/balde/app"
	"github.com/egermano/balde/core"
)

func TestStatusSeededBudget(t *testing.T) {
	store := newMemoryStore()
	mustCreateAccount(t, store, core.Account{Name: "Main", Type: core.AccountChecking})
	mustCreateBucket(t, store, core.Bucket{Name: "goals", Target: 50000, Balance: 20000, BudgetID: "default"})
	mustCreateTransaction(t, store, core.Transaction{
		Amount:      100000,
		Description: "salary",
		Date:        time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC),
		AccountID:   "acc-1",
	})

	a := app.New("default", store)

	status, err := a.Status()
	if err != nil {
		t.Fatalf("Status() error = %v, want nil", err)
	}

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("marshal status: %v", err)
	}

	want := `{"accounts":[{"id":"acc-1","name":"Main","type":"checking","balance":100000}],` +
		`"buckets":[{"id":"bkt-1","name":"goals","target":50000,"balance":20000,"budget_id":"default","archived":false,"fill_percent":40}],` +
		`"transactions":[{"id":"txn-1","amount":100000,"description":"salary","date":"2026-01-15T10:30:00Z","account_id":"acc-1","bucket_id":"","categorized":false}],` +
		`"rain":80000}`
	if string(data) != want {
		t.Errorf("Status() JSON =\n%s\nwant\n%s", data, want)
	}
}

func TestStatusEmptyBudget(t *testing.T) {
	a := app.New("default", newMemoryStore())

	status, err := a.Status()
	if err != nil {
		t.Fatalf("Status() error = %v, want nil", err)
	}

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("marshal status: %v", err)
	}

	want := `{"accounts":[],"buckets":[],"transactions":[],"rain":0}`
	if string(data) != want {
		t.Errorf("Status() JSON = %s, want %s", data, want)
	}
}
