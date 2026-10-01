package app_test

import (
	"testing"

	"github.com/egermano/balde/app"
	"github.com/egermano/balde/core"
)

func TestAllocateMovesRainToBucket(t *testing.T) {
	store := newMemoryStore()
	mustCreateAccount(t, store, core.Account{Name: "Main", Type: core.AccountChecking, Balance: 100000})
	mustCreateBucket(t, store, core.Bucket{Name: "goals", Target: 50000, BudgetID: "default"})
	a := app.New("default", store)

	if err := a.Allocate("bkt-1", 30000); err != nil {
		t.Fatalf("Allocate() error = %v, want nil", err)
	}

	status, err := a.Status()
	if err != nil {
		t.Fatalf("Status() error = %v, want nil", err)
	}
	if got := status.Buckets[0].Balance; got != 30000 {
		t.Errorf("bucket balance = %d, want 30000", got)
	}
	if got := status.Rain; got != 70000 {
		t.Errorf("rain = %d, want 70000", got)
	}
}

func TestAllocateArchivedBucketRejected(t *testing.T) {
	store := newMemoryStore()
	mustCreateBucket(t, store, core.Bucket{Name: "old", BudgetID: "default"})
	if err := store.DeleteBucket("bkt-1"); err != nil {
		t.Fatalf("archive bucket: %v", err)
	}
	a := app.New("default", store)

	if err := a.Allocate("bkt-1", 100); err == nil {
		t.Error("Allocate(archived) = nil error, want error")
	}
}

func TestAllocateUnknownBucketRejected(t *testing.T) {
	a := app.New("default", newMemoryStore())

	if err := a.Allocate("nope", 100); err == nil {
		t.Error("Allocate(unknown) = nil error, want error")
	}
}

func TestRainReturnsUnallocatedAmount(t *testing.T) {
	store := newMemoryStore()
	mustCreateAccount(t, store, core.Account{Name: "Main", Type: core.AccountChecking, Balance: 100000})
	mustCreateBucket(t, store, core.Bucket{Name: "goals", Target: 50000, Balance: 20000, BudgetID: "default"})
	a := app.New("default", store)

	rain, err := a.Rain()
	if err != nil {
		t.Fatalf("Rain() error = %v, want nil", err)
	}
	if rain != 80000 {
		t.Errorf("Rain() = %d, want 80000", rain)
	}
}
