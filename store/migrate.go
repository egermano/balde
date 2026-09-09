package store

import (
	"database/sql"
	"fmt"
)

const schema = `
CREATE TABLE IF NOT EXISTS accounts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	type TEXT NOT NULL,
	balance INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS buckets (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL CHECK (trim(name) <> ''),
	target INTEGER NOT NULL DEFAULT 0,
	balance INTEGER NOT NULL DEFAULT 0,
	budget_id TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS transactions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	amount INTEGER NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	date TEXT NOT NULL,
	account_id TEXT NOT NULL,
	bucket_id TEXT NOT NULL DEFAULT '',
	categorized INTEGER NOT NULL DEFAULT 0
);
`

func migrate(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(schema); err != nil {
		return err
	}

	var budgetID, normalizedName string
	err = tx.QueryRow(`
		SELECT budget_id, lower(trim(name))
		FROM buckets
		GROUP BY budget_id, lower(trim(name))
		HAVING COUNT(*) > 1
		LIMIT 1
	`).Scan(&budgetID, &normalizedName)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil {
		return fmt.Errorf("duplicate bucket name %q in budget %q", normalizedName, budgetID)
	}
	var emptyNameCount int
	if err := tx.QueryRow("SELECT COUNT(*) FROM buckets WHERE trim(name) = ''").Scan(&emptyNameCount); err != nil {
		return err
	}
	if emptyNameCount > 0 {
		return fmt.Errorf("bucket name is empty after trimming")
	}

	if _, err := tx.Exec("UPDATE buckets SET name = trim(name)"); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS buckets_budget_normalized_name_unique
		ON buckets (budget_id, lower(trim(name)))
	`); err != nil {
		return err
	}

	return tx.Commit()
}
