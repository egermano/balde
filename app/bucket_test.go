package app_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/egermano/balde/app"
	"github.com/egermano/balde/core"
)

func TestAddBucketReturnsDTO(t *testing.T) {
	store := newMemoryStore()
	a := app.New("default", store)

	bucket, err := a.AddBucket("vacation", 50000)
	if err != nil {
		t.Fatalf("AddBucket() error = %v, want nil", err)
	}

	want := app.Bucket{ID: "bkt-1", Name: "vacation", Target: 50000, Balance: 0, BudgetID: "default"}
	if bucket != want {
		t.Errorf("AddBucket() = %+v, want %+v", bucket, want)
	}

	buckets, err := a.ListBuckets()
	if err != nil {
		t.Fatalf("ListBuckets() error = %v, want nil", err)
	}
	if len(buckets) != 1 || buckets[0] != want {
		t.Errorf("ListBuckets() = %+v, want [%+v]", buckets, want)
	}
}

func TestAddBucketDuplicateNameRejected(t *testing.T) {
	a := app.New("default", newMemoryStore())

	if _, err := a.AddBucket("vacation", 50000); err != nil {
		t.Fatalf("first AddBucket() error = %v, want nil", err)
	}

	if _, err := a.AddBucket("Vacation", 1000); err == nil {
		t.Error("AddBucket(duplicate name) = nil error, want error")
	}
}

func TestAddBucketNinthBucketRejected(t *testing.T) {
	store := newMemoryStore()
	for i := 0; i < 8; i++ {
		mustCreateBucket(t, store, core.Bucket{Name: fmt.Sprintf("bucket-%d", i+1), BudgetID: "default"})
	}
	a := app.New("default", store)

	_, err := a.AddBucket("one-too-many", 100)
	if err == nil {
		t.Error("AddBucket(9th) = nil error, want error")
	}
}

func TestDeleteBucketArchivesAndCountsLinkedTransactions(t *testing.T) {
	store := newMemoryStore()
	mustCreateBucket(t, store, core.Bucket{Name: "comfort", Target: 10000, BudgetID: "default"})
	mustCreateAccount(t, store, core.Account{Name: "Main", Type: core.AccountChecking})
	mustCreateTransaction(t, store, core.Transaction{
		Amount: -2500, Description: "a", Date: time.Now(),
		AccountID: "acc-1", BucketID: "bkt-1", Categorized: true,
	})
	a := app.New("default", store)

	deleted, linked, err := a.DeleteBucket("bkt-1")
	if err != nil {
		t.Fatalf("DeleteBucket() error = %v, want nil", err)
	}
	if deleted.Name != "comfort" {
		t.Errorf("DeleteBucket() deleted = %+v, want comfort", deleted)
	}
	if linked != 1 {
		t.Errorf("DeleteBucket() linked = %d, want 1", linked)
	}

	buckets, err := a.ListBuckets()
	if err != nil {
		t.Fatalf("ListBuckets() error = %v, want nil", err)
	}
	if len(buckets) != 0 {
		t.Errorf("ListBuckets() after delete = %d buckets, want 0", len(buckets))
	}
}

func TestDeleteBucketUnknownIDReturnsError(t *testing.T) {
	a := app.New("default", newMemoryStore())

	if _, _, err := a.DeleteBucket("nope"); err == nil {
		t.Error("DeleteBucket(unknown) = nil error, want error")
	}
}

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
