package app

import "fmt"

// Status returns a snapshot of the whole budget: accounts, buckets,
// transactions and the unallocated rain.
func (a *App) Status() (Status, error) {
	accounts, err := a.store.ListAccounts()
	if err != nil {
		return Status{}, fmt.Errorf("status: %w", err)
	}

	buckets, err := a.store.ListBuckets()
	if err != nil {
		return Status{}, fmt.Errorf("status: %w", err)
	}

	transactions, err := a.store.ListTransactions()
	if err != nil {
		return Status{}, fmt.Errorf("status: %w", err)
	}

	rain, err := a.budget.Rain()
	if err != nil {
		return Status{}, fmt.Errorf("status: %w", err)
	}

	status := Status{
		Accounts:     make([]Account, 0, len(accounts)),
		Buckets:      make([]Bucket, 0, len(buckets)),
		Transactions: make([]Transaction, 0, len(transactions)),
		Rain:         rain,
	}
	for _, acc := range accounts {
		status.Accounts = append(status.Accounts, toAccountDTO(acc))
	}
	for _, bk := range buckets {
		status.Buckets = append(status.Buckets, a.bucketDTO(bk))
	}
	for _, t := range transactions {
		status.Transactions = append(status.Transactions, toTransactionDTO(t))
	}
	return status, nil
}

// Status is the budget snapshot DTO. Empty collections serialize as [].
type Status struct {
	Accounts     []Account     `json:"accounts"`
	Buckets      []Bucket      `json:"buckets"`
	Transactions []Transaction `json:"transactions"`
	Rain         int64         `json:"rain"`
}
