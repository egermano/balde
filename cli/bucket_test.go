package cli_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/egermano/balde/cli"
)

func executeBucketCommand(t *testing.T, args ...string) string {
	t.Helper()
	var output bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs(args)
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute %v: %v", args, err)
	}
	return output.String()
}

func TestBucketCmd_AddBucket(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	setupInitBudget(t)

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"bucket", "add", "housing", "50000"})
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("bucket add failed: %v", err)
	}
}

func TestBucketCmd_AddBucketInvalidTarget(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	setupInitBudget(t)

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"bucket", "add", "housing", "abc"})
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for invalid target")
	}
}

func TestBucketCmd_DeleteEmptyDefaultBucketWithoutPrompt(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	setupInitBudget(t)

	output := executeBucketCommand(t, "bucket", "delete", "2")
	if !strings.Contains(output, `Bucket "fixed costs" (id=2) archived`) {
		t.Fatalf("unexpected output: %q", output)
	}
	listed := executeBucketCommand(t, "view", "buckets", "--json")
	if strings.Contains(listed, `"fixed costs"`) {
		t.Fatalf("expected archived bucket excluded from active list: %s", listed)
	}
}

func TestBucketCmd_DeleteDefaultsToDenyWhenBucketHasBalance(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	setupInitBudget(t)
	executeBucketCommand(t, "account", "add", "checking", "checking", "10000")
	executeBucketCommand(t, "allocate", "1000", "1")

	var output bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"bucket", "delete", "1"})
	cmd.SetIn(strings.NewReader("\n"))
	cmd.SetOut(&output)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("delete bucket: %v", err)
	}
	if !strings.Contains(output.String(), "balance=1000") || !strings.Contains(output.String(), "Bucket not archived") {
		t.Fatalf("unexpected confirmation output: %q", output.String())
	}
	if listed := executeBucketCommand(t, "view", "buckets", "--json"); !strings.Contains(listed, `"financial freedom"`) {
		t.Fatal("expected denied archive to preserve active bucket")
	}
}

func TestBucketCmd_DeleteForceArchivesBucketWithTransactions(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	setupInitBudget(t)
	executeBucketCommand(t, "account", "add", "checking", "checking", "10000")
	executeBucketCommand(t, "transaction", "add", "-100", "snack", "1", "1")

	output := executeBucketCommand(t, "bucket", "delete", "1", "--force")
	if !strings.Contains(output, `Bucket "financial freedom" (id=1) archived`) {
		t.Fatalf("unexpected output: %q", output)
	}
	txs := executeBucketCommand(t, "view", "transactions", "--json")
	if !strings.Contains(txs, `"BucketID": "1"`) {
		t.Fatalf("expected transaction link preserved: %s", txs)
	}
}
