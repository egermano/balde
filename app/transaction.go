package app

import "time"

// ListTransactions returns all transactions as DTOs. Empty result is a
// non-nil empty slice.
func (a *App) ListTransactions() ([]Transaction, error) {
	transactions, err := a.store.ListTransactions()
	if err != nil {
		return nil, err
	}
	out := make([]Transaction, 0, len(transactions))
	for _, t := range transactions {
		out = append(out, toTransactionDTO(t))
	}
	return out, nil
}

// AddTransaction creates a transaction and returns its DTO. A non-empty
// bucketID categorizes it.
func (a *App) AddTransaction(amount int64, description string, date time.Time, accountID, bucketID string) (Transaction, error) {
	t, err := a.budget.AddTransaction(amount, description, date, accountID, bucketID)
	if err != nil {
		return Transaction{}, err
	}
	return toTransactionDTO(t), nil
}
