package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/egermano/balde/cli"
)

func TestViewCmd_BucketsJSON(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	setupInitBudget(t)

	root := cli.NewRootCmd()
	root.SetArgs([]string{"bucket", "add", "housing", "50000"})
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	root.Execute()

	root = cli.NewRootCmd()
	root.SetArgs([]string{"bucket", "add", "food", "30000"})
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
	if len(buckets) != 2 {
		t.Errorf("expected 2 buckets, got %d", len(buckets))
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
