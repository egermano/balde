package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/egermano/balde/cli"
)

func TestPluginCmd_ListShowsInstalledPlugin(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BALDE_EXPERIMENTAL", "1")

	pluginsDir := filepath.Join(dir, ".balde", "plugins")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	lock := `{"plugins":[{"name":"vacation","version":"0.1.0","source":"github.com/user/balde-plugin-vacation","sha":"abc","artifact_hash":"sha256:x","capabilities":[{"type":"command","name":"vacation"}],"permissions":{"read":["accounts"],"write":["allocate"]}}]}`
	if err := os.WriteFile(filepath.Join(pluginsDir, "lock.json"), []byte(lock), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"plugin", "list"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("plugin list error = %v, want nil", err)
	}
	if !strings.Contains(buf.String(), "vacation") ||
		!strings.Contains(buf.String(), "0.1.0") ||
		!strings.Contains(buf.String(), "github.com/user/balde-plugin-vacation") {
		t.Errorf("list output = %q, want plugin name/version/source", buf.String())
	}
}

func TestPluginCmd_InstallBadSourceErrors(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BALDE_EXPERIMENTAL", "1")

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"plugin", "install", "/nonexistent/repo"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("plugin install bad source = nil error, want error")
	}
	if !strings.Contains(err.Error(), "clone") {
		t.Errorf("error = %q, want it to mention clone", err)
	}
}

func TestPluginCmd_RemoveUnknownErrors(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BALDE_EXPERIMENTAL", "1")

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"plugin", "remove", "ghost"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("plugin remove unknown = nil error, want error")
	}
	if !strings.Contains(err.Error(), "not installed") {
		t.Errorf("error = %q, want it to mention not installed", err)
	}
}
