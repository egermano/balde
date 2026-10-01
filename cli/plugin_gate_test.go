package cli_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/egermano/balde/cli"
)

func TestPluginCmd_RefusedWithoutOptIn(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BALDE_EXPERIMENTAL", "")

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"plugin", "list"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("plugin list without opt-in = nil error, want refusal")
	}
	if !strings.Contains(err.Error(), "experimental") {
		t.Errorf("error = %q, want it to mention experimental", err)
	}
	if !strings.Contains(err.Error(), "BALDE_EXPERIMENTAL") {
		t.Errorf("error = %q, want it to explain how to opt in", err)
	}
}

func TestPluginCmd_FlagOptIn(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BALDE_EXPERIMENTAL", "")

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"plugin", "list", "--experimental"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("plugin list --experimental error = %v, want nil", err)
	}
	if !strings.Contains(buf.String(), "No plugins installed") {
		t.Errorf("output = %q, want no-plugins notice", buf.String())
	}
}

func TestPluginCmd_EnvOptIn(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BALDE_EXPERIMENTAL", "1")

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"plugin", "list"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("plugin list with BALDE_EXPERIMENTAL=1 error = %v, want nil", err)
	}
}
