package cli_test

import (
	"os"
	"testing"

	"github.com/egermano/balde/cli"
)

func setupInitBudget(t *testing.T) {
	t.Helper()
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"init"})
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init: %v", err)
	}
}

func TestAccountCmd_AddAccount(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	setupInitBudget(t)

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"account", "add", "checking", "checking", "100000"})
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("account add failed: %v", err)
	}
}

func TestAccountCmd_AddAccountInvalidBalance(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	setupInitBudget(t)

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"account", "add", "checking", "checking", "notanumber"})
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for invalid balance")
	}
}
