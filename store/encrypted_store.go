package store

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/egermano/balde/core"
	_ "modernc.org/sqlite"
)

type EncryptedSQLiteStore struct {
	db       *sql.DB
	encPath  string
	password string
}

func NewEncryptedSQLiteStore(encPath, password string) (*EncryptedSQLiteStore, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("open in-memory db: %w", err)
	}

	s := &EncryptedSQLiteStore{
		db:       db,
		encPath:  encPath,
		password: password,
	}

	if _, err := os.Stat(encPath); err == nil {
		if err := s.loadEncrypted(); err != nil {
			db.Close()
			return nil, fmt.Errorf("load encrypted db: %w", err)
		}
	} else {
		if err := s.migrate(); err != nil {
			db.Close()
			return nil, fmt.Errorf("migrate: %w", err)
		}
	}

	return s, nil
}

func (s *EncryptedSQLiteStore) Close() error {
	defer s.db.Close()
	return s.saveEncrypted()
}

func (s *EncryptedSQLiteStore) loadEncrypted() error {
	data, err := os.ReadFile(s.encPath)
	if err != nil {
		return fmt.Errorf("read encrypted file: %w", err)
	}

	ciphertext, salt, nonce, err := DeserializeEncrypted(data)
	if err != nil {
		return fmt.Errorf("deserialize: %w", err)
	}

	plaintext, err := Decrypt(ciphertext, salt, nonce, s.password)
	if err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}

	dump := string(plaintext)

	for _, stmt := range splitStatements(dump) {
		if stmt == "" {
			continue
		}
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("restore from dump: %w\nstatement: %s", err, stmt)
		}
	}

	return nil
}

func splitStatements(dump string) []string {
	var stmts []string
	var current string
	inString := false
	escaped := false

	for i := 0; i < len(dump); i++ {
		c := dump[i]

		if escaped {
			current = current + string(c)
			escaped = false
			continue
		}

		if c == '\\' {
			current = current + string(c)
			escaped = true
			continue
		}

		if c == '\'' {
			inString = !inString
			current = current + string(c)
			continue
		}

		if c == ';' && !inString {
			stmts = append(stmts, current)
			current = ""
			continue
		}

		current = current + string(c)
	}

	if current != "" {
		stmts = append(stmts, current)
	}

	return stmts
}

func (s *EncryptedSQLiteStore) saveEncrypted() error {
	var dump []byte

	tables, err := s.getTables()
	if err != nil {
		return fmt.Errorf("get tables: %w", err)
	}

	for _, table := range tables {
		if table == "sqlite_sequence" {
			continue
		}

		rows, err := s.db.Query("SELECT sql FROM sqlite_master WHERE type='table' AND name = ?", table)
		if err != nil {
			return fmt.Errorf("query schema for %s: %w", table, err)
		}

		for rows.Next() {
			var sqlStr string
			if err := rows.Scan(&sqlStr); err != nil {
				return err
			}
			dump = append(dump, sqlStr...)
			dump = append(dump, 59) // semicolon
		}
		rows.Close()

		rows, err = s.db.Query("SELECT * FROM " + table)
		if err != nil {
			return fmt.Errorf("query table %s: %w", table, err)
		}

		columns, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("get columns for %s: %w", table, err)
		}

		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range columns {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return err
			}

			dump = append(dump, "INSERT INTO "+table+" VALUES ("...)

			for i, v := range values {
				if i > 0 {
					dump = append(dump, 44) // comma
				}
				switch val := v.(type) {
				case nil:
					dump = append(dump, "NULL"...)
				case []byte:
					dump = append(dump, 39) // quote
					dump = append(dump, escapeSQLString(string(val))...)
					dump = append(dump, 39)
				case int:
					dump = append(dump, fmt.Sprintf("%d", val)...)
				case int64:
					dump = append(dump, fmt.Sprintf("%d", val)...)
				default:
					dump = append(dump, 39) // quote
					dump = append(dump, escapeSQLString(fmt.Sprintf("%v", val))...)
					dump = append(dump, 39)
				}
			}
			dump = append(dump, 41, 59) // close paren, semicolon
		}
		rows.Close()
	}

	ciphertext, salt, nonce, err := Encrypt(dump, s.password)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	serialized, err := SerializeEncrypted(ciphertext, salt, nonce)
	if err != nil {
		return fmt.Errorf("serialize: %w", err)
	}

	if err := os.WriteFile(s.encPath, serialized, 0600); err != nil {
		return fmt.Errorf("write encrypted file: %w", err)
	}

	return nil
}

func (s *EncryptedSQLiteStore) getTables() ([]string, error) {
	rows, err := s.db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, nil
}

func escapeSQLString(s string) string {
	result := make([]byte, 0, len(s)*2)
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\'':
			result = append(result, "\\'"...)
		case '\\':
			result = append(result, "\\\\"...)
		default:
			result = append(result, s[i])
		}
	}
	return string(result)
}

func (s *EncryptedSQLiteStore) migrate() error {
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

func (s *EncryptedSQLiteStore) CreateAccount(a core.Account) error {
	_, err := s.db.Exec(
		"INSERT INTO accounts (name, type, balance) VALUES (?, ?, ?)",
		a.Name, a.Type, a.Balance,
	)
	return err
}

func (s *EncryptedSQLiteStore) GetAccount(id string) (core.Account, error) {
	var a core.Account
	err := s.db.QueryRow(
		"SELECT id, name, type, balance FROM accounts WHERE id = ?", id,
	).Scan(&a.ID, &a.Name, &a.Type, &a.Balance)
	if err != nil {
		return core.Account{}, fmt.Errorf("account not found: %s", id)
	}
	return a, nil
}

func (s *EncryptedSQLiteStore) ListAccounts() ([]core.Account, error) {
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

func (s *EncryptedSQLiteStore) UpdateAccount(a core.Account) error {
	_, err := s.db.Exec(
		"UPDATE accounts SET name = ?, type = ?, balance = ? WHERE id = ?",
		a.Name, a.Type, a.Balance, a.ID,
	)
	return err
}

func (s *EncryptedSQLiteStore) CreateBucket(b core.Bucket) error {
	_, err := s.db.Exec(
		"INSERT INTO buckets (name, target, balance, budget_id) VALUES (?, ?, ?, ?)",
		b.Name, b.Target, b.Balance, b.BudgetID,
	)
	return err
}

func (s *EncryptedSQLiteStore) GetBucket(id string) (core.Bucket, error) {
	var b core.Bucket
	err := s.db.QueryRow(
		"SELECT id, name, target, balance, budget_id FROM buckets WHERE id = ?", id,
	).Scan(&b.ID, &b.Name, &b.Target, &b.Balance, &b.BudgetID)
	if err != nil {
		return core.Bucket{}, fmt.Errorf("bucket not found: %s", id)
	}
	return b, nil
}

func (s *EncryptedSQLiteStore) ListBuckets() ([]core.Bucket, error) {
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

func (s *EncryptedSQLiteStore) UpdateBucket(b core.Bucket) error {
	_, err := s.db.Exec(
		"UPDATE buckets SET name = ?, target = ?, balance = ?, budget_id = ? WHERE id = ?",
		b.Name, b.Target, b.Balance, b.BudgetID, b.ID,
	)
	return err
}

func (s *EncryptedSQLiteStore) DeleteBucket(id string) error {
	_, err := s.db.Exec("DELETE FROM buckets WHERE id = ?", id)
	return err
}

func (s *EncryptedSQLiteStore) CreateTransaction(t core.Transaction) error {
	_, err := s.db.Exec(
		"INSERT INTO transactions (amount, description, date, account_id, bucket_id, categorized) VALUES (?, ?, ?, ?, ?, ?)",
		t.Amount, t.Description, t.Date.Format("2006-01-02T15:04:05Z"), t.AccountID, t.BucketID, t.Categorized,
	)
	return err
}

func (s *EncryptedSQLiteStore) GetTransaction(id string) (core.Transaction, error) {
	var tx core.Transaction
	var dateStr string
	var categorized int
	err := s.db.QueryRow(
		"SELECT id, amount, description, date, account_id, bucket_id, categorized FROM transactions WHERE id = ?", id,
	).Scan(&tx.ID, &tx.Amount, &tx.Description, &dateStr, &tx.AccountID, &tx.BucketID, &categorized)
	if err != nil {
		return core.Transaction{}, fmt.Errorf("transaction not found: %s", id)
	}
	tx.Date, _ = parseDate(dateStr)
	tx.Categorized = categorized == 1
	return tx, nil
}

func (s *EncryptedSQLiteStore) ListTransactions() ([]core.Transaction, error) {
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
		tx.Date, _ = parseDate(dateStr)
		tx.Categorized = categorized == 1
		txs = append(txs, tx)
	}
	return txs, nil
}

func (s *EncryptedSQLiteStore) UpdateTransaction(t core.Transaction) error {
	_, err := s.db.Exec(
		"UPDATE transactions SET amount = ?, description = ?, date = ?, account_id = ?, bucket_id = ?, categorized = ? WHERE id = ?",
		t.Amount, t.Description, t.Date.Format("2006-01-02T15:04:05Z"), t.AccountID, t.BucketID, t.Categorized, t.ID,
	)
	return err
}
