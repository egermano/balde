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
