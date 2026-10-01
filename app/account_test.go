package app_test

import (
	"testing"

	"github.com/egermano/balde/app"
	"github.com/egermano/balde/core"
)

func TestAddAccountPersistsAndReturnsDTO(t *testing.T) {
	a := app.New("default", newMemoryStore())

	acc, err := a.AddAccount("Main", core.AccountChecking, 100000)
	if err != nil {
		t.Fatalf("AddAccount() error = %v, want nil", err)
	}

	want := app.Account{ID: "acc-1", Name: "Main", Type: core.AccountChecking, Balance: 100000}
	if acc != want {
		t.Errorf("AddAccount() = %+v, want %+v", acc, want)
	}

	status, err := a.Status()
	if err != nil {
		t.Fatalf("Status() error = %v, want nil", err)
	}
	if len(status.Accounts) != 1 || status.Accounts[0] != want {
		t.Errorf("Status().Accounts = %+v, want [%+v]", status.Accounts, want)
	}
	if status.Rain != 100000 {
		t.Errorf("Status().Rain = %d, want 100000", status.Rain)
	}
}
