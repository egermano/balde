package cli

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/egermano/balde/core"
	"github.com/egermano/balde/store"
)

func TestEncryptCmd_PlainDBToEncrypted(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create a plain DB
	cmd := newInitCmd()
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Verify it's plain
	config, _ := store.ReadConfig("balde.db")
	if config.Encrypted {
		t.Fatalf("initial DB should be plain")
	}

	// Encrypt it
	encryptCmd := newEncryptCmd()
	encryptCmd.Flags().Set("password", "newsecret")
	if err := encryptCmd.RunE(encryptCmd, []string{}); err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// Verify it's now encrypted
	config, _ = store.ReadConfig("balde.db")
	if !config.Encrypted {
		t.Fatalf("DB should be encrypted after encrypt command")
	}

	// Verify backup exists
	if _, err := os.Stat("balde.db.backup"); os.IsNotExist(err) {
		t.Fatalf("backup file should be created")
	}

	// Verify can open with new password
	s, err := store.OpenStore("balde.db", "newsecret", "", config)
	if err != nil {
		t.Fatalf("failed to open encrypted DB: %v", err)
	}
	defer s.Close()

	// Verify buckets are still there
	buckets, _ := s.ListBuckets()
	if len(buckets) != 6 {
		t.Fatalf("expected 6 buckets, got %d", len(buckets))
	}
}

func TestEncryptCmd_AlreadyEncrypted(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create encrypted DB
	cmd := newInitCmd()
	cmd.Flags().Set("password", "original")
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Try to encrypt again
	encryptCmd := newEncryptCmd()
	encryptCmd.Flags().Set("password", "newpassword")
	err := encryptCmd.RunE(encryptCmd, []string{})
	if err == nil || err.Error() != "database is already encrypted" {
		t.Fatalf("expected 'database is already encrypted' error, got: %v", err)
	}
}

func TestEncryptCmd_NoPassword(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create plain DB
	cmd := newInitCmd()
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Try to encrypt without password
	encryptCmd := newEncryptCmd()
	err := encryptCmd.RunE(encryptCmd, []string{})
	if err == nil {
		t.Fatalf("expected error for missing password")
	}
}

func TestEncryptCmd_BackupNameCollision(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create plain DB
	cmd := newInitCmd()
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Create a backup file
	os.WriteFile("balde.db.backup", []byte("old backup"), 0644)

	// Encrypt (should handle collision)
	encryptCmd := newEncryptCmd()
	encryptCmd.Flags().Set("password", "newsecret")
	if err := encryptCmd.RunE(encryptCmd, []string{}); err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// Verify DB is encrypted
	config, _ := store.ReadConfig("balde.db")
	if !config.Encrypted {
		t.Fatalf("DB should be encrypted")
	}
}

// TestEncryptCmd_MigrationPreservesStateWithArchivedBucket is the regression
// test for #51: encrypt must migrate budgets whose transactions reference
// archived buckets, and it must copy rows 1:1 — same IDs, same balances —
// instead of re-applying transaction deltas onto remapped IDs (which
// silently corrupted balances and landed deltas on wrong buckets).
func TestEncryptCmd_MigrationPreservesStateWithArchivedBucket(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)
	t.Setenv("BALDE_PASSWORD", "")

	cmd := newInitCmd()
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Seed a realistic budget: income, expense, allocation, and a bucket
	// that gets archived while a transaction still references it.
	plain, err := store.NewSQLiteStore("balde.db")
	if err != nil {
		t.Fatal(err)
	}
	b := core.NewBudget("default", plain)
	if _, err := b.AddAccount("checking", "checking", 100000); err != nil {
		t.Fatal(err)
	}
	if _, err := b.AddBucket("oldcar", 30000); err != nil {
		t.Fatal(err)
	}
	buckets, err := plain.ListBuckets()
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
	if err := plain.UpdateBucket(core.Bucket{ID: oldcarID, Name: "oldcar", BudgetID: "default", Archived: true}); err != nil {
		t.Fatal(err)
	}

	before := snapshotForMigration(t, plain)
	plain.Close()

	encryptCmd := newEncryptCmd()
	encryptCmd.Flags().Set("password", "newsecret")
	if err := encryptCmd.RunE(encryptCmd, []string{}); err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// A failed migration used to leave a poisoned balde.db.encrypted behind;
	// a successful one must not either (it becomes balde.db).
	if _, err := os.Stat("balde.db.encrypted"); err == nil {
		t.Error("balde.db.encrypted leftover after successful migration")
	}

	enc, err := store.OpenStore("balde.db", "newsecret", "", store.Config{Encrypted: true})
	if err != nil {
		t.Fatalf("open encrypted store: %v", err)
	}
	defer enc.Close()
	after := snapshotForMigration(t, enc)
	if before != after {
		t.Errorf("state changed across encryption:\nbefore:\n%s\nafter:\n%s", before, after)
	}

	// The command prints "Session created for 30 minutes" — the session must
	// actually work for the final database path, without a password.
	var rainOut bytes.Buffer
	root := NewRootCmd()
	root.SetArgs([]string{"rain"})
	root.SetOut(&rainOut)
	root.SetErr(&rainOut)
	if err := root.Execute(); err != nil {
		t.Errorf("rain after encrypt should use the session, got: %v", err)
	}
}

func snapshotForMigration(t *testing.T, s core.Store) string {
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
		fmt.Fprintf(&sb, "acc %s %s %s %d\n", a.ID, a.Name, a.Type, a.Balance)
	}
	for _, bk := range buckets {
		fmt.Fprintf(&sb, "bkt %s %s %d %d archived=%v\n", bk.ID, bk.Name, bk.Target, bk.Balance, bk.Archived)
	}
	for _, tx := range transactions {
		fmt.Fprintf(&sb, "txn %s %d %s acc=%s bkt=%s\n", tx.ID, tx.Amount, tx.Description, tx.AccountID, tx.BucketID)
	}
	return sb.String()
}
