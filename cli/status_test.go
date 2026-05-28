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
	if !ok || len(buckets) != 1 {
		t.Errorf("expected 1 bucket, got %v", status["buckets"])
	}
}
