package core

import "time"

type AccountType string

const (
	AccountChecking AccountType = "checking"
	AccountSavings  AccountType = "savings"
	AccountCredit   AccountType = "credit"
)

type Account struct {
	ID      string
	Name    string
	Type    AccountType
	Balance int64
}

type Bucket struct {
	ID       string
	Name     string
	Target   int64
	Balance  int64
	BudgetID string
}

type Transaction struct {
	ID          string
	Amount      int64
	Description string
	Date        time.Time
	AccountID   string
	BucketID    string
	Categorized bool
}
