package app


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
