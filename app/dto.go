package app

import (
	"time"

	"github.com/egermano/balde/core"
)

// DTOs define the stable wire schema. Field names and JSON tags are part of
// the public contract (CLI --json output, plugin protocol) and must not
// change casually.

type Account struct {
	ID      string           `json:"id"`
	Name    string           `json:"name"`
	Type    core.AccountType `json:"type"`
	Balance int64            `json:"balance"`
}

type Bucket struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Target      int64   `json:"target"`
	Balance     int64   `json:"balance"`
	BudgetID    string  `json:"budget_id"`
	Archived    bool    `json:"archived"`
	FillPercent float64 `json:"fill_percent"`
}

type Transaction struct {
	ID          string    `json:"id"`
	Amount      int64     `json:"amount"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	AccountID   string    `json:"account_id"`
	BucketID    string    `json:"bucket_id"`
	Categorized bool      `json:"categorized"`
}

func toAccountDTO(a core.Account) Account {
	return Account{
		ID:      a.ID,
		Name:    a.Name,
		Type:    a.Type,
		Balance: a.Balance,
	}
}

func toBucketDTO(b core.Bucket) Bucket {
	return Bucket{
		ID:       b.ID,
		Name:     b.Name,
		Target:   b.Target,
		Balance:  b.Balance,
		BudgetID: b.BudgetID,
		Archived: b.Archived,
	}
}

func toTransactionDTO(t core.Transaction) Transaction {
	return Transaction{
		ID:          t.ID,
		Amount:      t.Amount,
		Description: t.Description,
		Date:        t.Date,
		AccountID:   t.AccountID,
		BucketID:    t.BucketID,
		Categorized: t.Categorized,
	}
}
