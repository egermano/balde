package core

import (
	"fmt"
	"time"
)

type Budget struct {
	ID    string
	store Store
}

func NewBudget(id string, store Store) *Budget {
	return &Budget{ID: id, store: store}
}

func (b *Budget) AddAccount(name string, accountType AccountType, initialBalance int64) (Account, error) {
	a := Account{
		Name:    name,
		Type:    accountType,
		Balance: initialBalance,
	}
	if err := b.store.CreateAccount(a); err != nil {
		return Account{}, fmt.Errorf("add account: %w", err)
	}

	accounts, err := b.store.ListAccounts()
	if err != nil {
		return Account{}, fmt.Errorf("add account: %w", err)
	}
	return accounts[len(accounts)-1], nil
}

func (b *Budget) AddBucket(name string, target int64) (Bucket, error) {
	buckets, err := b.store.ListBuckets()
	if err != nil {
		return Bucket{}, fmt.Errorf("add bucket: %w", err)
	}
	if len(buckets) >= 8 {
		return Bucket{}, fmt.Errorf("add bucket: maximum of 8 buckets exceeded")
	}

	bk := Bucket{
		Name:     name,
		Target:   target,
		Balance:  0,
		BudgetID: b.ID,
	}
	if err := b.store.CreateBucket(bk); err != nil {
		return Bucket{}, fmt.Errorf("add bucket: %w", err)
	}

	buckets, err = b.store.ListBuckets()
	if err != nil {
		return Bucket{}, fmt.Errorf("add bucket: %w", err)
	}
	return buckets[len(buckets)-1], nil
}

func (b *Budget) AddTransaction(amount int64, description string, date time.Time, accountID, bucketID string) (Transaction, error) {
	t := Transaction{
		Amount:      amount,
		Description: description,
		Date:        date,
		AccountID:   accountID,
		BucketID:    bucketID,
	}
	if err := b.store.CreateTransaction(t); err != nil {
		return Transaction{}, fmt.Errorf("add transaction: %w", err)
	}

	// Retrieve the actual transaction with the assigned ID
	transactions, err := b.store.ListTransactions()
	if err != nil {
		return Transaction{}, fmt.Errorf("add transaction: %w", err)
	}
	return transactions[len(transactions)-1], nil
}

func (b *Budget) Allocate(bucketID string, amount int64) error {
	bk, err := b.store.GetBucket(bucketID)
	if err != nil {
		return fmt.Errorf("allocate: %w", err)
	}

	bk.Balance += amount
	if err := b.store.UpdateBucket(bk); err != nil {
		return fmt.Errorf("allocate: %w", err)
	}
	return nil
}

func (b *Budget) Rain() (int64, error) {
	accounts, err := b.store.ListAccounts()
	if err != nil {
		return 0, fmt.Errorf("rain: %w", err)
	}

	buckets, err := b.store.ListBuckets()
	if err != nil {
		return 0, fmt.Errorf("rain: %w", err)
	}

	var totalAccounts int64
	for _, a := range accounts {
		totalAccounts += a.Balance
	}

	var totalBuckets int64
	for _, bk := range buckets {
		totalBuckets += bk.Balance
	}

	return totalAccounts - totalBuckets, nil
}

func (b *Budget) CalculateFillPercentage(bucket Bucket) float64 {
	// Handle zero target to avoid division by zero
	if bucket.Target == 0 {
		return 0.0
	}

	// Calculate fill percentage: (balance / target) * 100
	percent := float64(bucket.Balance) / float64(bucket.Target) * 100
	return percent
}
