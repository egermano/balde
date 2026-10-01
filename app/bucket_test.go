package app_test

import (
	"testing"

	"github.com/egermano/balde/app"
	"github.com/egermano/balde/core"
)

func TestListBucketsReturnsDTOs(t *testing.T) {
	store := newMemoryStore()
	mustCreateBucket(t, store, core.Bucket{Name: "goals", Target: 50000, Balance: 20000, BudgetID: "default"})
	a := app.New("default", store)

	buckets, err := a.ListBuckets()
	if err != nil {
		t.Fatalf("ListBuckets() error = %v, want nil", err)
	}

	if len(buckets) != 1 {
		t.Fatalf("ListBuckets() returned %d buckets, want 1", len(buckets))
	}
	want := app.Bucket{ID: "bkt-1", Name: "goals", Target: 50000, Balance: 20000, BudgetID: "default"}
	if buckets[0] != want {
		t.Errorf("ListBuckets()[0] = %+v, want %+v", buckets[0], want)
	}
}

func TestListBucketsExcludesArchived(t *testing.T) {
	store := newMemoryStore()
	mustCreateBucket(t, store, core.Bucket{Name: "goals", Target: 50000, Balance: 20000, BudgetID: "default"})
	mustCreateBucket(t, store, core.Bucket{Name: "old", Target: 1000, Balance: 0, BudgetID: "default"})
	if err := store.DeleteBucket("bkt-2"); err != nil {
		t.Fatalf("archive bucket: %v", err)
	}
	a := app.New("default", store)

	buckets, err := a.ListBuckets()
	if err != nil {
		t.Fatalf("ListBuckets() error = %v, want nil", err)
	}

	if len(buckets) != 1 || buckets[0].Name != "goals" {
		t.Errorf("ListBuckets() = %+v, want only goals", buckets)
	}
}

func TestListBucketsEmptyBudgetReturnsEmptySlice(t *testing.T) {
	a := app.New("default", newMemoryStore())

	buckets, err := a.ListBuckets()
	if err != nil {
		t.Fatalf("ListBuckets() error = %v, want nil", err)
	}

	if buckets == nil {
		t.Error("ListBuckets() = nil, want empty non-nil slice")
	}
	if len(buckets) != 0 {
		t.Errorf("ListBuckets() returned %d buckets, want 0", len(buckets))
	}
}
