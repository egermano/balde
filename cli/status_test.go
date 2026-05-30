package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/egermano/balde/cli"
)

func TestStatusCmd_JSON(t *testing.T) {
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
	root.SetArgs([]string{"allocate", "30000", "1"})
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	root.Execute()

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"status", "--json"})
	cmd.SetOut(&buf)
	cmd.SetErr(os.Stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("status failed: %v", err)
	}

	var status map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &status); err != nil {
		t.Fatalf("parse JSON: %v\noutput: %s", err, buf.String())
	}

	if status["rain"] == nil {
		t.Error("expected rain field in status")
	}
	accounts, ok := status["accounts"].([]interface{})
	if !ok || len(accounts) != 1 {
		t.Errorf("expected 1 account, got %v", status["accounts"])
	}
	buckets, ok := status["buckets"].([]interface{})
	if !ok || len(buckets) != 7 {
		t.Errorf("expected 7 buckets (6 default + 1 added), got %v", status["buckets"])
	}
}

func TestStatusCmd_JSON_EmptyBudget(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	setupInitBudget(t)

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"status", "--json"})
	cmd.SetOut(&buf)
	cmd.SetErr(os.Stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("status failed: %v", err)
	}

	var status map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &status); err != nil {
		t.Fatalf("parse JSON: %v\noutput: %s", err, buf.String())
	}

	// Test that empty collections are arrays, not null
	accounts, ok := status["accounts"].([]interface{})
	if !ok {
		t.Errorf("expected accounts to be array, got %T (value: %v)", status["accounts"], status["accounts"])
	} else if len(accounts) != 0 {
		t.Errorf("expected 0 accounts in fresh budget, got %d", len(accounts))
	}

	transactions, ok := status["transactions"].([]interface{})
	if !ok {
		t.Errorf("expected transactions to be array, got %T (value: %v)", status["transactions"], status["transactions"])
	} else if len(transactions) != 0 {
		t.Errorf("expected 0 transactions in fresh budget, got %d", len(transactions))
	}

	// Buckets should have the 6 default buckets
	buckets, ok := status["buckets"].([]interface{})
	if !ok {
		t.Errorf("expected buckets to be array, got %T (value: %v)", status["buckets"], status["buckets"])
	} else if len(buckets) != 6 {
		t.Errorf("expected 6 default buckets, got %d", len(buckets))
	}

	if status["rain"] == nil {
		t.Error("expected rain field in status")
	}
}
