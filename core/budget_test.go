package core_test

import (
	"testing"
	"time"

	"github.com/egermano/balde/core"
)

func TestBudget_AddTransaction_ReturnsTransactionWithID(t *testing.T) {
	store := NewMemoryStore()
	budget := core.NewBudget("b1", store)

	// Add an account first
	acc, err := budget.AddAccount("checking", core.AccountChecking, 100000)
	if err != nil {
		t.Fatalf("add account: %v", err)
	}

	// Add a bucket first
	bkt, err := budget.AddBucket("food", 50000)
	if err != nil {
		t.Fatalf("add bucket: %v", err)
	}

	// Test: Add transaction should return transaction with non-empty ID
	tx, err := budget.AddTransaction(-3500, "Coffee", time.Now(), acc.ID, bkt.ID)
	if err != nil {
		t.Fatalf("add transaction: %v", err)
	}

	if tx.ID == "" {
		t.Error("expected transaction ID to be non-empty, got empty string")
	}

	// Additional checks
	if tx.Amount != -3500 {
		t.Errorf("expected amount -3500, got %d", tx.Amount)
	}
	if tx.Description != "Coffee" {
		t.Errorf("expected description 'Coffee', got %s", tx.Description)
	}
	if tx.AccountID != acc.ID {
		t.Errorf("expected account ID %s, got %s", acc.ID, tx.AccountID)
	}
	if tx.BucketID != bkt.ID {
		t.Errorf("expected bucket ID %s, got %s", bkt.ID, tx.BucketID)
	}

	// Test: Transaction should be retrievable by ID
	retrievedTx, err := store.GetTransaction(tx.ID)
	if err != nil {
		t.Errorf("failed to retrieve transaction by ID %s: %v", tx.ID, err)
	}

	if retrievedTx.ID != tx.ID {
		t.Errorf("retrieved transaction ID %s doesn't match original ID %s", retrievedTx.ID, tx.ID)
	}
	if retrievedTx.Amount != tx.Amount {
		t.Errorf("retrieved transaction amount %d doesn't match original amount %d", retrievedTx.Amount, tx.Amount)
	}

	// Test: Add another transaction to ensure multiple work
	tx2, err := budget.AddTransaction(-2000, "Lunch", time.Now(), acc.ID, bkt.ID)
	if err != nil {
		t.Fatalf("add second transaction: %v", err)
	}

	if tx2.ID == "" {
		t.Error("expected second transaction ID to be non-empty, got empty string")
	}

	// Verify second transaction is also retrievable
	retrievedTx2, err := store.GetTransaction(tx2.ID)
	if err != nil {
		t.Errorf("failed to retrieve second transaction by ID %s: %v", tx2.ID, err)
	}

	if retrievedTx2.ID != tx2.ID {
		t.Errorf("retrieved second transaction ID %s doesn't match original ID %s", retrievedTx2.ID, tx2.ID)
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
