package app_test

import (
	"testing"

	"github.com/egermano/balde/app"
	"github.com/egermano/balde/core"
)

func TestListAccountsReturnsDTOs(t *testing.T) {
	store := newMemoryStore()
	mustCreateAccount(t, store, core.Account{Name: "Main", Type: core.AccountChecking, Balance: 100000})
	mustCreateAccount(t, store, core.Account{Name: "Wallet", Type: core.AccountSavings, Balance: 5000})
	a := app.New("default", store)

	accounts, err := a.ListAccounts()
	if err != nil {
		t.Fatalf("ListAccounts() error = %v, want nil", err)
	}

	if len(accounts) != 2 {
		t.Fatalf("ListAccounts() returned %d accounts, want 2", len(accounts))
	}
	byID := map[string]app.Account{}
	for _, acc := range accounts {
		byID[acc.ID] = acc
	}
	main := byID["acc-1"]
	if main.Name != "Main" || main.Type != core.AccountChecking || main.Balance != 100000 {
		t.Errorf("acc-1 = %+v, want Main/checking/100000", main)
	}
	wallet := byID["acc-2"]
	if wallet.Name != "Wallet" || wallet.Type != core.AccountSavings || wallet.Balance != 5000 {
		t.Errorf("acc-2 = %+v, want Wallet/savings/5000", wallet)
	}
}

func TestListAccountsEmptyBudgetReturnsEmptySlice(t *testing.T) {
	a := app.New("default", newMemoryStore())

	accounts, err := a.ListAccounts()
	if err != nil {
		t.Fatalf("ListAccounts() error = %v, want nil", err)
	}

	if accounts == nil {
		t.Error("ListAccounts() = nil, want empty non-nil slice")
	}
	if len(accounts) != 0 {
		t.Errorf("ListAccounts() returned %d accounts, want 0", len(accounts))
	}
}
