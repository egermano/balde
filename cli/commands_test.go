package cli_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/egermano/balde/cli"
)

func setupTransaction(t *testing.T) {
	t.Helper()
	for _, args := range [][]string{
		{"account", "add", "checking", "checking", "100000"},
		{"bucket", "add", "housing", "50000"},
		{"transaction", "add", "-3500", "Coffee", "1", "1"},
	} {
		cmd := cli.NewRootCmd()
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("execute %v: %v", args, err)
		}
	}
}

func TestTransactionCmd_AddTransaction(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	setupInitBudget(t)

	root := cli.NewRootCmd()
	root.SetArgs([]string{"account", "add", "checking", "checking", "100000"})
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	if err := root.Execute(); err != nil {
		t.Fatalf("account add: %v", err)
	}

	root = cli.NewRootCmd()
	root.SetArgs([]string{"bucket", "add", "housing", "50000"})
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	if err := root.Execute(); err != nil {
		t.Fatalf("bucket add: %v", err)
	}

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"transaction", "add", "-50000", "rent", "1", "1"})
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("transaction add failed: %v", err)
	}
}

func TestTransactionCmd_Delete_DefaultsToDeny(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	setupInitBudget(t)
	setupTransaction(t)

	var output bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"transaction", "delete", "1"})
	cmd.SetIn(strings.NewReader("\n"))
	cmd.SetOut(&output)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("delete transaction: %v", err)
	}
	if !strings.Contains(output.String(), "Coffee (-3500)") {
		t.Errorf("expected transaction details, got %q", output.String())
	}

	output.Reset()
	view := cli.NewRootCmd()
	view.SetArgs([]string{"view", "transactions", "--json"})
	view.SetOut(&output)
	if err := view.Execute(); err != nil {
		t.Fatalf("view transactions: %v", err)
	}
	if !strings.Contains(output.String(), `"Description": "Coffee"`) {
		t.Error("expected denied deletion to preserve transaction")
	}
}

func TestTransactionCmd_Delete_ForceDeletes(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	setupInitBudget(t)
	setupTransaction(t)

	var output bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"transaction", "delete", "1", "--force"})
	cmd.SetOut(&output)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("delete transaction: %v", err)
	}
	if !strings.Contains(output.String(), "Transaction deleted: Coffee (-3500)") {
		t.Errorf("unexpected output: %q", output.String())
	}

	var statusOutput bytes.Buffer
	status := cli.NewRootCmd()
	status.SetArgs([]string{"status", "--json"})
	status.SetOut(&statusOutput)
	if err := status.Execute(); err != nil {
		t.Fatalf("status: %v", err)
	}
	if !strings.Contains(statusOutput.String(), `"Balance": 100000`) {
		t.Errorf("expected account balance to be restored, got %q", statusOutput.String())
	}
	if !strings.Contains(statusOutput.String(), `"Balance": 0`) {
		t.Errorf("expected bucket balance to be restored, got %q", statusOutput.String())
	}
}

func TestTransactionCmd_Delete_ConfirmedDeletes(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	setupInitBudget(t)
	setupTransaction(t)

	var output bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"transaction", "delete", "1"})
	cmd.SetIn(strings.NewReader("y\n"))
	cmd.SetOut(&output)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("delete transaction: %v", err)
	}
	if !strings.Contains(output.String(), "Transaction deleted: Coffee (-3500)") {
		t.Errorf("unexpected output: %q", output.String())
	}
}

func TestAllocateCmd(t *testing.T) {
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

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"allocate", "50000", "1"})
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("allocate failed: %v", err)
	}
}

func TestRainCmd(t *testing.T) {
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
	root.SetArgs([]string{"allocate", "50000", "1"})
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	root.Execute()

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"rain"})
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("rain failed: %v", err)
	}
}
