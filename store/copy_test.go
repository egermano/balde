package store_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/egermano/balde/core"
	"github.com/egermano/balde/store"
)

// seedMigrationFixture builds a plain budget with an archived bucket that
// still has a transaction attached, non-zero balances everywhere, and
// non-sequential IDs — the shape that broke the old copy loop.
func seedMigrationFixture(t *testing.T, s *store.SQLiteStore) {
	t.Helper()
	b := core.NewBudget("default", s)
	if _, err := b.AddAccount("checking", "checking", 100000); err != nil {
		t.Fatal(err)
	}
	if _, err := b.AddBucket("oldcar", 30000); err != nil {
		t.Fatal(err)
	}
	buckets, err := s.ListBuckets()
	if err != nil {
		t.Fatal(err)
	}
	var goalsID, oldcarID string
	for _, bk := range buckets {
		switch bk.Name {
		case "goals":
			goalsID = bk.ID
		case "oldcar":
			oldcarID = bk.ID
		}
	}
	if _, err := b.AddTransaction(250000, "salary", time.Now(), "1", goalsID); err != nil {
		t.Fatal(err)
	}
	if _, err := b.AddTransaction(-1000, "toll", time.Now(), "1", oldcarID); err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateBucket(core.Bucket{ID: oldcarID, Name: "oldcar", BudgetID: "default", Archived: true}); err != nil {
		t.Fatal(err)
	}
}

func dumpAll(t *testing.T, s core.Store) string {
	t.Helper()
	accounts, err := s.ListAccounts()
	if err != nil {
		t.Fatal(err)
	}
	buckets, err := s.ListAllBuckets()
	if err != nil {
		t.Fatal(err)
	}
	transactions, err := s.ListTransactions()
	if err != nil {
		t.Fatal(err)
	}
	var sb strings.Builder
	for _, a := range accounts {
		fmt.Fprintf(&sb, "acc %s %s %d\n", a.ID, a.Name, a.Balance)
	}
	for _, bk := range buckets {
		fmt.Fprintf(&sb, "bkt %s %s %d %d archived=%v\n", bk.ID, bk.Name, bk.Target, bk.Balance, bk.Archived)
	}
	for _, tx := range transactions {
		fmt.Fprintf(&sb, "txn %s %d acc=%s bkt=%s\n", tx.ID, tx.Amount, tx.AccountID, tx.BucketID)
	}
	return sb.String()
}

// TestCopyBudget_PreservesIDsBalancesAndArchived copies a plain budget into
// a fresh encrypted store and requires the destination to be an exact row
// mirror: identical IDs (so transaction references stay valid), identical
// balances (no re-applied transaction deltas), archived buckets included.
func TestCopyBudget_PreservesIDsBalancesAndArchived(t *testing.T) {
	srcPath := filepath.Join(t.TempDir(), "src.db")
	src, err := store.NewSQLiteStore(srcPath)
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	seedMigrationFixture(t, src)
	want := dumpAll(t, src)

	encPath := filepath.Join(t.TempDir(), "enc.db")
	dst, err := store.NewEncryptedSQLiteStore(encPath, "pass")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CopyBudget(src, dst); err != nil {
		t.Fatalf("copy: %v", err)
	}
	if err := dst.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := store.NewEncryptedSQLiteStore(encPath, "pass")
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if got := dumpAll(t, reopened); got != want {
		t.Errorf("copy mismatch:\nwant:\n%s\ngot:\n%s", want, got)
	}
}

// TestEncryptedStore_DiscardDoesNotPersist verifies the failure-path
// primitive: Discard closes the store without persisting anything, so a
// failed migration can never leave a poisoned partial file behind.
func TestEncryptedStore_DiscardDoesNotPersist(t *testing.T) {
	encPath := filepath.Join(t.TempDir(), "enc.db")
	s, err := store.NewEncryptedSQLiteStore(encPath, "pass")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateAccount(core.Account{Name: "checking", Type: core.AccountChecking, Balance: 100}); err != nil {
		t.Fatal(err)
	}
	if err := s.Discard(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(encPath); !os.IsNotExist(err) {
		t.Fatalf("discarded store must not persist %s (stat err=%v)", encPath, err)
	}

	reopened, err := store.NewEncryptedSQLiteStore(encPath, "pass")
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	accounts, err := reopened.ListAccounts()
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 0 {
		t.Errorf("accounts after discard = %v, want none", accounts)
	}
}
