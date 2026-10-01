package app

// Allocate moves an amount from rain into a bucket. Amounts are cents;
// negative amounts move money back to rain.
func (a *App) Allocate(bucketID string, amount int64) error {
	return a.budget.Allocate(bucketID, amount)
}

// Rain returns the unallocated amount: sum of account balances minus sum of
// bucket balances.
func (a *App) Rain() (int64, error) {
	return a.budget.Rain()
}
