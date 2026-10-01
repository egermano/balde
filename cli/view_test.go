package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/egermano/balde/cli"
	"github.com/egermano/balde/store"
)

func TestViewCmd_AccountsTextUsesBudgetConfigAndDirectory(t *testing.T) {
	budgetDir := t.TempDir()
	otherDir := t.TempDir()
	if err := os.Chdir(otherDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(otherDir) })

	initCmd := cli.NewRootCmd()
	initCmd.SetArgs([]string{"init", "--dir", budgetDir})
	initCmd.SetIn(strings.NewReader("n\n"))
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init: %v", err)
	}

	dbPath := filepath.Join(budgetDir, "balde.db")
	config, err := store.ReadConfig(dbPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	config.CurrencySymbol = "R$"
	config.DecimalSeparator = ","
	config.ThousandsSeparator = "."
	if err := store.WriteConfig(dbPath, config); err != nil {
		t.Fatalf("write config: %v", err)
	}

	addCmd := cli.NewRootCmd()
	addCmd.SetArgs([]string{"account", "add", "My Checking", "checking", "123456", "--dir", budgetDir})
	if err := addCmd.Execute(); err != nil {
		t.Fatalf("account add: %v", err)
	}

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"view", "accounts", "--dir", budgetDir})
	cmd.SetOut(&buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("view accounts: %v", err)
	}

	want := "1\tMy Checking\tchecking\tR$1.234,56\n"
	if got := buf.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestViewCmd_AccountsJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	setupInitBudget(t)

	addCmd := cli.NewRootCmd()
	addCmd.SetArgs([]string{"account", "add", "Savings", "savings", "500000"})
	if err := addCmd.Execute(); err != nil {
		t.Fatalf("account add: %v", err)
	}

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"view", "accounts", "--json"})
	cmd.SetOut(&buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("view accounts: %v", err)
	}

	var accounts []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &accounts); err != nil {
		t.Fatalf("parse JSON: %v\noutput: %s", err, buf.String())
	}
	if len(accounts) != 1 {
		t.Fatalf("accounts count = %d, want 1", len(accounts))
	}

	wantKeys := []string{"id", "name", "type", "balance"}
	got := accounts[0]
	for _, key := range wantKeys {
		if _, ok := got[key]; !ok {
			t.Errorf("account JSON missing key %q, got keys: %v", key, mapKeys(got))
		}
	}
	if _, ok := got["ID"]; ok {
		t.Errorf("account JSON has untagged key %q, got keys: %v", "ID", mapKeys(got))
	}
	if got["id"] != "1" || got["name"] != "Savings" || got["type"] != "savings" || got["balance"] != float64(500000) {
		t.Errorf("account = %v", got)
	}
}

func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func TestViewCmd_AccountsEmptyJSONIsArray(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	setupInitBudget(t)

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"view", "accounts", "--json"})
	cmd.SetOut(&buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("view accounts: %v", err)
	}

	if got := strings.TrimSpace(buf.String()); got != "[]" {
		t.Fatalf("output = %q, want []", got)
	}
}

func TestViewCmd_BucketsJSON(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	setupInitBudget(t)

	root := cli.NewRootCmd()
	root.SetArgs([]string{"bucket", "add", "extra1", "50000"})
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	root.Execute()

	root = cli.NewRootCmd()
	root.SetArgs([]string{"bucket", "add", "extra2", "30000"})
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	root.Execute()

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"view", "buckets", "--json"})
	cmd.SetOut(&buf)
	cmd.SetErr(os.Stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("view buckets failed: %v", err)
	}

	var buckets []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &buckets); err != nil {
		t.Fatalf("parse JSON: %v\noutput: %s", err, buf.String())
	}
	if len(buckets) != 8 {
		t.Errorf("expected 8 buckets (6 default + 2 added), got %d", len(buckets))
	}
}

func TestViewCmd_TransactionsJSON(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	setupInitBudget(t)

	root := cli.NewRootCmd()
	root.SetArgs([]string{"account", "add", "checking", "checking", "100000"})
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	root.Execute()

	root = cli.NewRootCmd()
	root.SetArgs([]string{"bucket", "add", "housing", "50000"})
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	root.Execute()

	root = cli.NewRootCmd()
	root.SetArgs([]string{"transaction", "add", "-50000", "rent", "1", "1"})
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	root.Execute()

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"view", "transactions", "--json"})
	cmd.SetOut(&buf)
	cmd.SetErr(os.Stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("view transactions failed: %v", err)
	}

	var txs []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &txs); err != nil {
		t.Fatalf("parse JSON: %v\noutput: %s", err, buf.String())
	}
	if len(txs) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(txs))
	}
}
