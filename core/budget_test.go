package core_test

import (
	"fmt"
	"testing"

	"github.com/egermano/balde/core"
)

func TestRain_NoAccountsOrBuckets(t *testing.T) {
	store := NewMemoryStore()
	budget := core.NewBudget("b1", store)

	rain, err := budget.Rain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rain != 0 {
		t.Errorf("expected rain=0, got %d", rain)
	}
}

func TestRain_WithIncomeOnly(t *testing.T) {
	store := NewMemoryStore()
	budget := core.NewBudget("b1", store)

	_, err := budget.AddAccount("checking", core.AccountChecking, 100000)
	if err != nil {
		t.Fatalf("add account: %v", err)
	}

	rain, err := budget.Rain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rain != 100000 {
		t.Errorf("expected rain=100000, got %d", rain)
	}
}

func TestRain_AfterAllocation(t *testing.T) {
	store := NewMemoryStore()
	budget := core.NewBudget("b1", store)

	acc, _ := budget.AddAccount("checking", core.AccountChecking, 100000)
	bkt, _ := budget.AddBucket("housing", 50000)

	budget.Allocate(bkt.ID, 50000)

	rain, err := budget.Rain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rain != 50000 {
		t.Errorf("expected rain=50000, got %d", rain)
	}

	_ = acc
}

func TestRain_FullyAllocated(t *testing.T) {
	store := NewMemoryStore()
	budget := core.NewBudget("b1", store)

	budget.AddAccount("checking", core.AccountChecking, 100000)
	bkt, _ := budget.AddBucket("housing", 50000)
	bkt2, _ := budget.AddBucket("food", 50000)

	budget.Allocate(bkt.ID, 50000)
	budget.Allocate(bkt2.ID, 50000)

	rain, err := budget.Rain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rain != 0 {
		t.Errorf("expected rain=0, got %d", rain)
	}
}

func TestRain_OverAllocated(t *testing.T) {
	store := NewMemoryStore()
	budget := core.NewBudget("b1", store)

	budget.AddAccount("checking", core.AccountChecking, 100000)
	bkt, _ := budget.AddBucket("housing", 80000)

	budget.Allocate(bkt.ID, 120000)

	rain, err := budget.Rain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rain != -20000 {
		t.Errorf("expected rain=-20000, got %d", rain)
	}
}

func TestAddBucket_ExceedsMax(t *testing.T) {
	store := NewMemoryStore()
	budget := core.NewBudget("b1", store)

	for i := 0; i < 8; i++ {
		_, err := budget.AddBucket(fmt.Sprintf("bucket-%d", i), 10000)
		if err != nil {
			t.Fatalf("bucket %d should succeed: %v", i, err)
		}
	}

	_, err := budget.AddBucket("bucket-9", 10000)
	if err == nil {
		t.Fatal("expected error when adding 9th bucket")
	}
}

func TestAllocate_UpdatesBucketBalance(t *testing.T) {
	store := NewMemoryStore()
	budget := core.NewBudget("b1", store)

	budget.AddAccount("checking", core.AccountChecking, 100000)
	bkt, _ := budget.AddBucket("housing", 50000)

	budget.Allocate(bkt.ID, 30000)

	updated, err := store.GetBucket(bkt.ID)
	if err != nil {
		t.Fatalf("get bucket: %v", err)
	}
	if updated.Balance != 30000 {
		t.Errorf("expected bucket balance=30000, got %d", updated.Balance)
	}
}

func TestBudget_CalculateFillPercentage(t *testing.T) {
	store := NewMemoryStore()
	budget := core.NewBudget("b1", store)

	// Test case 1: Normal bucket with target and balance
	t.Run("normal bucket", func(t *testing.T) {
		bkt := core.Bucket{
			ID:      "1",
			Name:    "housing",
			Target:  50000,
			Balance: 25000,
		}
		
		percent := budget.CalculateFillPercentage(bkt)
		expected := 50.0 // 25000/50000 * 100
		
		if percent != expected {
			t.Errorf("expected fill percentage %.2f, got %.2f", expected, percent)
		}
	})

	// Test case 2: Bucket with zero target (should not crash)
	t.Run("zero target bucket", func(t *testing.T) {
		bkt := core.Bucket{
			ID:      "2",
			Name:    "financial freedom",
			Target:  0,
			Balance: 0,
		}
		
		// This should not panic or crash with division by zero
		percent := budget.CalculateFillPercentage(bkt)
		
		// According to the issue, this should show "Not set" or "-" instead of NaN
		if percent != 0.0 { // Or some other sensible default
			t.Errorf("expected fill percentage for zero target to be 0 or special value, got %.2f", percent)
		}
	})

	// Test case 3: Bucket with balance but zero target
	t.Run("zero target with balance", func(t *testing.T) {
		bkt := core.Bucket{
			ID:      "3",
			Name:    "fixed costs",
			Target:  0,
			Balance: 100000,
		}
		
		percent := budget.CalculateFillPercentage(bkt)
		// Should handle gracefully without division by zero
		if percent < 0 {
			t.Errorf("fill percentage should not be negative for zero target, got %.2f", percent)
		}
	})
}