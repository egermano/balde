package cli_test

import (
	"os"
	"testing"

	"github.com/egermano/balde/cli"
)

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
