package cli_test

import (
	"os"
	"testing"

	"github.com/egermano/balde/cli"
)

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
