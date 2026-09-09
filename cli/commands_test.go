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

func TestAccountAddSupportsBudgetDirectory(t *testing.T) {
	budgetDir := t.TempDir()
	otherDir := t.TempDir()
	t.Cleanup(func() { _ = os.Chdir(otherDir) })

	if err := os.Chdir(otherDir); err != nil {
		t.Fatal(err)
	}

	initCmd := cli.NewRootCmd()
	initCmd.SetArgs([]string{"init", "--dir", budgetDir})
	initCmd.SetIn(bytes.NewBufferString("n\n"))
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init: %v", err)
	}

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"account", "add", "checking", "checking", "100000", "--dir", budgetDir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("account add with --dir: %v", err)
	}

	if _, err := os.Stat(budgetDir + "/balde.db"); err != nil {
		t.Fatalf("budget database: %v", err)
	}
	if _, err := os.Stat(otherDir + "/balde.db"); !os.IsNotExist(err) {
		t.Fatalf("unexpected database in working directory: %v", err)
	}
}

func TestBudgetCommandsSupportDirectory(t *testing.T) {
	budgetDir := t.TempDir()
	otherDir := t.TempDir()
	t.Cleanup(func() { _ = os.Chdir(otherDir) })
	if err := os.Chdir(otherDir); err != nil {
		t.Fatal(err)
	}

	initCmd := cli.NewRootCmd()
	initCmd.SetArgs([]string{"init", "--dir", budgetDir})
	initCmd.SetIn(bytes.NewBufferString("n\n"))
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init: %v", err)
	}

	commands := [][]string{
		{"account", "add", "checking", "checking", "100000", "--dir", budgetDir},
		{"bucket", "add", "housing", "50000", "--dir", budgetDir},
		{"transaction", "add", "-1000", "rent", "1", "7", "--dir", budgetDir},
		{"allocate", "50000", "7", "--dir", budgetDir},
		{"rain", "--dir", budgetDir},
		{"view", "buckets", "--dir", budgetDir},
		{"view", "transactions", "--dir", budgetDir},
		{"status", "--dir", budgetDir},
	}

	for _, args := range commands {
		cmd := cli.NewRootCmd()
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("%v: %v", args, err)
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

func TestAllocateCmdDeclinesOverAllocationByDefault(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	setupInitBudget(t)

	runCommand(t, "account", "add", "checking", "checking", "50000")

	var output bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"allocate", "200000", "1"})
	cmd.SetIn(strings.NewReader("\n"))
	cmd.SetOut(&output)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("allocate: %v", err)
	}

	want := "Warning: You have 50000 cents available, but trying to allocate 200000 cents.\nThis will result in negative rain. Continue? (y/N) "
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}

	var rainOutput bytes.Buffer
	rainCmd := cli.NewRootCmd()
	rainCmd.SetArgs([]string{"rain"})
	rainCmd.SetOut(&rainOutput)
	if err := rainCmd.Execute(); err != nil {
		t.Fatalf("rain: %v", err)
	}
	if !strings.Contains(rainOutput.String(), "50000 cents") {
		t.Fatalf("rain output = %q, want unchanged rain", rainOutput.String())
	}
}

func TestAllocateCmdAcceptsOverAllocationConfirmationCaseInsensitively(t *testing.T) {
	for _, answer := range []string{"y", "YES", "Yes"} {
		t.Run(answer, func(t *testing.T) {
			dir := t.TempDir()
			os.Chdir(dir)
			setupInitBudget(t)
			runCommand(t, "account", "add", "checking", "checking", "50000")

			cmd := cli.NewRootCmd()
			cmd.SetArgs([]string{"allocate", "200000", "1"})
			cmd.SetIn(strings.NewReader(answer + "\n"))
			if err := cmd.Execute(); err != nil {
				t.Fatalf("allocate: %v", err)
			}

			var rainOutput bytes.Buffer
			rainCmd := cli.NewRootCmd()
			rainCmd.SetArgs([]string{"rain"})
			rainCmd.SetOut(&rainOutput)
			if err := rainCmd.Execute(); err != nil {
				t.Fatalf("rain: %v", err)
			}
			if !strings.Contains(rainOutput.String(), "-150000 cents") {
				t.Fatalf("rain output = %q, want negative rain", rainOutput.String())
			}
		})
	}
}

func TestAllocateCmdForceSkipsOverAllocationPrompt(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	setupInitBudget(t)
	runCommand(t, "account", "add", "checking", "checking", "50000")

	var output bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"allocate", "200000", "1", "--force"})
	cmd.SetOut(&output)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("allocate: %v", err)
	}
	if strings.Contains(output.String(), "Continue?") {
		t.Fatalf("output = %q, want no prompt", output.String())
	}

	var rainOutput bytes.Buffer
	rainCmd := cli.NewRootCmd()
	rainCmd.SetArgs([]string{"rain"})
	rainCmd.SetOut(&rainOutput)
	if err := rainCmd.Execute(); err != nil {
		t.Fatalf("rain: %v", err)
	}
	if !strings.Contains(rainOutput.String(), "-150000 cents") {
		t.Fatalf("rain output = %q, want forced negative rain", rainOutput.String())
	}
}

func TestAllocateCmdRejectsInvalidBucketBeforePrompting(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	setupInitBudget(t)
	runCommand(t, "account", "add", "checking", "checking", "50000")

	var output bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"allocate", "200000", "missing"})
	cmd.SilenceUsage = true
	cmd.SetOut(&output)
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "bucket not found") {
		t.Fatalf("error = %v, want bucket not found", err)
	}
	if output.Len() != 0 {
		t.Fatalf("output = %q, want no prompt", output.String())
	}
}

func runCommand(t *testing.T, args ...string) {
	t.Helper()
	cmd := cli.NewRootCmd()
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("%v: %v", args, err)
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
