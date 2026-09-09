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

func TestBudget_AddTransaction_AppliesCategorizedExpenseToBalances(t *testing.T) {
	store := NewMemoryStore()
	budget := core.NewBudget("b1", store)

	account, err := budget.AddAccount("checking", core.AccountChecking, 10000)
	if err != nil {
		t.Fatalf("add account: %v", err)
	}
	bucket, err := budget.AddBucket("food", 5000)
	if err != nil {
		t.Fatalf("add bucket: %v", err)
	}

	if _, err := budget.AddTransaction(-2500, "groceries", time.Now(), account.ID, bucket.ID); err != nil {
		t.Fatalf("add transaction: %v", err)
	}

	account, err = store.GetAccount(account.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if account.Balance != 7500 {
		t.Errorf("expected account balance 7500, got %d", account.Balance)
	}

	bucket, err = store.GetBucket(bucket.ID)
	if err != nil {
		t.Fatalf("get bucket: %v", err)
	}
	if bucket.Balance != -2500 {
		t.Errorf("expected bucket balance -2500, got %d", bucket.Balance)
	}
}

func TestBudget_DeleteTransaction_ReversesCategorizedBalances(t *testing.T) {
	store := NewMemoryStore()
	budget := core.NewBudget("b1", store)
	account, _ := budget.AddAccount("checking", core.AccountChecking, 10000)
	bucket, _ := budget.AddBucket("food", 5000)
	tx, err := budget.AddTransaction(-2500, "groceries", time.Now(), account.ID, bucket.ID)
	if err != nil {
		t.Fatalf("add transaction: %v", err)
	}

	deleted, err := budget.DeleteTransaction(tx.ID)
	if err != nil {
		t.Fatalf("delete transaction: %v", err)
	}
	if deleted.ID != tx.ID || deleted.Description != "groceries" {
		t.Errorf("unexpected deleted transaction: %+v", deleted)
	}

	account, _ = store.GetAccount(account.ID)
	bucket, _ = store.GetBucket(bucket.ID)
	if account.Balance != 10000 {
		t.Errorf("expected account balance 10000, got %d", account.Balance)
	}
	if bucket.Balance != 0 {
		t.Errorf("expected bucket balance 0, got %d", bucket.Balance)
	}
	if _, err := store.GetTransaction(tx.ID); err == nil {
		t.Error("expected transaction to be deleted")
	}
}

func TestBudget_AddTransaction_UncategorizedOnlyAffectsAccount(t *testing.T) {
	store := NewMemoryStore()
	budget := core.NewBudget("b1", store)
	account, _ := budget.AddAccount("checking", core.AccountChecking, 10000)
	bucket, _ := budget.AddBucket("food", 5000)

	tx, err := budget.AddTransaction(2500, "income", time.Now(), account.ID, "")
	if err != nil {
		t.Fatalf("add transaction: %v", err)
	}
	if tx.Categorized {
		t.Error("expected transaction without bucket to be uncategorized")
	}
	account, _ = store.GetAccount(account.ID)
	bucket, _ = store.GetBucket(bucket.ID)
	if account.Balance != 12500 {
		t.Errorf("expected account balance 12500, got %d", account.Balance)
	}
	if bucket.Balance != 0 {
		t.Errorf("expected unchanged bucket balance 0, got %d", bucket.Balance)
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
