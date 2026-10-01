package app

import "github.com/egermano/balde/core"

// AddAccount creates an account and returns its DTO.
func (a *App) AddAccount(name string, accountType core.AccountType, initialBalance int64) (Account, error) {
	acc, err := a.budget.AddAccount(name, accountType, initialBalance)
	if err != nil {
		return Account{}, err
	}
	return toAccountDTO(acc), nil
}

// ListAccounts returns all accounts as DTOs. Empty result is a non-nil
// empty slice.
func (a *App) ListAccounts() ([]Account, error) {
	accounts, err := a.store.ListAccounts()
	if err != nil {
		return nil, err
	}
	out := make([]Account, 0, len(accounts))
	for _, acc := range accounts {
		out = append(out, toAccountDTO(acc))
	}
	return out, nil
}
