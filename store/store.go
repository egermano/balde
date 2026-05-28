package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/egermano/balde/core"
	_ "modernc.org/sqlite"
)

type Store interface {
	Close() error
	CreateAccount(a core.Account) error
	GetAccount(id string) (core.Account, error)
	ListAccounts() ([]core.Account, error)
	UpdateAccount(a core.Account) error
	CreateBucket(b core.Bucket) error
	GetBucket(id string) (core.Bucket, error)
	ListBuckets() ([]core.Bucket, error)
	UpdateBucket(b core.Bucket) error
	DeleteBucket(id string) error
	CreateTransaction(t core.Transaction) error
	GetTransaction(id string) (core.Transaction, error)
	ListTransactions() ([]core.Transaction, error)
	UpdateTransaction(t core.Transaction) error
}

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	s := &SQLiteStore{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS accounts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		balance INTEGER NOT NULL DEFAULT 0
	);
	CREATE TABLE IF NOT EXISTS buckets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
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
	_, err := s.db.Exec(schema)
	return err
}

func (s *SQLiteStore) CreateAccount(a core.Account) error {
	_, err := s.db.Exec(
		"INSERT INTO accounts (name, type, balance) VALUES (?, ?, ?)",
		a.Name, a.Type, a.Balance,
	)
	return err
}

func (s *SQLiteStore) GetAccount(id string) (core.Account, error) {
	var a core.Account
	err := s.db.QueryRow(
		"SELECT id, name, type, balance FROM accounts WHERE id = ?", id,
	).Scan(&a.ID, &a.Name, &a.Type, &a.Balance)
	if err != nil {
		return core.Account{}, fmt.Errorf("account not found: %s", id)
	}
	return a, nil
}

func (s *SQLiteStore) ListAccounts() ([]core.Account, error) {
	rows, err := s.db.Query("SELECT id, name, type, balance FROM accounts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []core.Account
	for rows.Next() {
		var a core.Account
		if err := rows.Scan(&a.ID, &a.Name, &a.Type, &a.Balance); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func (s *SQLiteStore) UpdateAccount(a core.Account) error {
	_, err := s.db.Exec(
		"UPDATE accounts SET name = ?, type = ?, balance = ? WHERE id = ?",
		a.Name, a.Type, a.Balance, a.ID,
	)
	return err
}

func (s *SQLiteStore) CreateBucket(b core.Bucket) error {
	_, err := s.db.Exec(
		"INSERT INTO buckets (name, target, balance, budget_id) VALUES (?, ?, ?, ?)",
		b.Name, b.Target, b.Balance, b.BudgetID,
	)
	return err
}

func (s *SQLiteStore) GetBucket(id string) (core.Bucket, error) {
	var b core.Bucket
	err := s.db.QueryRow(
		"SELECT id, name, target, balance, budget_id FROM buckets WHERE id = ?", id,
	).Scan(&b.ID, &b.Name, &b.Target, &b.Balance, &b.BudgetID)
	if err != nil {
		return core.Bucket{}, fmt.Errorf("bucket not found: %s", id)
	}
	return b, nil
}

func (s *SQLiteStore) ListBuckets() ([]core.Bucket, error) {
	rows, err := s.db.Query("SELECT id, name, target, balance, budget_id FROM buckets")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buckets []core.Bucket
	for rows.Next() {
		var b core.Bucket
		if err := rows.Scan(&b.ID, &b.Name, &b.Target, &b.Balance, &b.BudgetID); err != nil {
			return nil, err
		}
		buckets = append(buckets, b)
	}
	return buckets, nil
}

func (s *SQLiteStore) UpdateBucket(b core.Bucket) error {
	_, err := s.db.Exec(
		"UPDATE buckets SET name = ?, target = ?, balance = ?, budget_id = ? WHERE id = ?",
		b.Name, b.Target, b.Balance, b.BudgetID, b.ID,
	)
	return err
}

func (s *SQLiteStore) DeleteBucket(id string) error {
	_, err := s.db.Exec("DELETE FROM buckets WHERE id = ?", id)
	return err
}

func (s *SQLiteStore) CreateTransaction(t core.Transaction) error {
	_, err := s.db.Exec(
		"INSERT INTO transactions (amount, description, date, account_id, bucket_id, categorized) VALUES (?, ?, ?, ?, ?, ?)",
		t.Amount, t.Description, t.Date.Format("2006-01-02T15:04:05Z"), t.AccountID, t.BucketID, t.Categorized,
	)
	return err
}

func (s *SQLiteStore) GetTransaction(id string) (core.Transaction, error) {
	var tx core.Transaction
	var dateStr string
	var categorized int
	err := s.db.QueryRow(
		"SELECT id, amount, description, date, account_id, bucket_id, categorized FROM transactions WHERE id = ?", id,
	).Scan(&tx.ID, &tx.Amount, &tx.Description, &dateStr, &tx.AccountID, &tx.BucketID, &categorized)
	if err != nil {
		return core.Transaction{}, fmt.Errorf("transaction not found: %s", id)
	}
	tx.Date, _ = time.Parse(time.RFC3339, dateStr)
	tx.Categorized = categorized == 1
	return tx, nil
}

func (s *SQLiteStore) ListTransactions() ([]core.Transaction, error) {
	rows, err := s.db.Query("SELECT id, amount, description, date, account_id, bucket_id, categorized FROM transactions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []core.Transaction
	for rows.Next() {
		var tx core.Transaction
		var dateStr string
		var categorized int
		if err := rows.Scan(&tx.ID, &tx.Amount, &tx.Description, &dateStr, &tx.AccountID, &tx.BucketID, &categorized); err != nil {
			return nil, err
		}
		tx.Date, _ = time.Parse(time.RFC3339, dateStr)
		tx.Categorized = categorized == 1
		txs = append(txs, tx)
	}
	return txs, nil
}

func (s *SQLiteStore) UpdateTransaction(t core.Transaction) error {
	_, err := s.db.Exec(
		"UPDATE transactions SET amount = ?, description = ?, date = ?, account_id = ?, bucket_id = ?, categorized = ? WHERE id = ?",
		t.Amount, t.Description, t.Date.Format("2006-01-02T15:04:05Z"), t.AccountID, t.BucketID, t.Categorized, t.ID,
	)
	return err
}

func OpenStore(dbPath string, password string, envPassword string, config Config) (Store, error) {
	if !config.Encrypted {
		return NewSQLiteStore(dbPath)
	}

	if password == "" && envPassword != "" {
		password = envPassword
	}

	sessionPath := GetSessionPath(dbPath)

	if password == "" {
		session, err := ReadSession(sessionPath)
		if err != nil {
			return nil, fmt.Errorf("no password provided and no valid session: %w", err)
		}

		store, err := NewEncryptedSQLiteStore(dbPath, session.Password)
		if err != nil {
			return nil, fmt.Errorf("open encrypted store: %w", err)
		}

		if err := RenewSession(sessionPath, 30*time.Minute); err != nil {
			return nil, fmt.Errorf("renew session: %w", err)
		}

		return store, nil
	}

	store, err := NewEncryptedSQLiteStore(dbPath, password)
	if err != nil {
		return nil, fmt.Errorf("open encrypted store: %w", err)
	}

	session := Session{
		Password:     password,
		LastAccessed: time.Now(),
	}

	if err := WriteSession(sessionPath, session, 30*time.Minute); err != nil {
		return nil, fmt.Errorf("write session: %w", err)
	}

	return store, nil
}
