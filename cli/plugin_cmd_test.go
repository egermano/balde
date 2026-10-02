package cli_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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
	setupInitBudget(t)

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
	setupInitBudget(t)

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

func TestInstalledPluginCommandRunsWithOptIn(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BALDE_EXPERIMENTAL", "1")
	setupInitBudget(t)

	pluginPath := filepath.Join(dir, ".balde", "plugins", "src", "hello", "bin", "plugin")
	// This static protocol plugin responds to host initialize (id 1), then
	// command/execute (id 2). It proves Cobra registration + runtime wiring.
	script := `#!/bin/sh
while IFS= read -r line; do
  case "$line" in
    *initialize*) printf '%s\n' '{"jsonrpc":"2.0","id":1,"result":{"name":"hello","protocol":1}}' ;;
    *command/execute*) printf '%s\n' '{"jsonrpc":"2.0","id":2,"result":{"text":"hello from installed plugin","json":{"ok":true}}}'; exit 0 ;;
  esac
done
`
	if err := os.MkdirAll(filepath.Dir(pluginPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pluginPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(pluginPath)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	lock := `{"plugins":[{"name":"hello","version":"0.1.0","source":"fixture","sha":"x","artifact_hash":"sha256:` + hex.EncodeToString(sum[:]) + `","run":"bin/plugin","capabilities":[{"type":"command","name":"hello","description":"fixture"}],"permissions":{}}]}`
	if err := os.WriteFile(filepath.Join(dir, ".balde", "plugins", "lock.json"), []byte(lock), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"hello", "plan", "Japan"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("installed plugin command error = %v", err)
	}
	if !strings.Contains(buf.String(), "hello from installed plugin") {
		t.Errorf("output = %q, want plugin text", buf.String())
	}
}

func TestInstalledPluginCommandJSONOutput(t *testing.T) {
	// The behavior is covered in the command above for text. This test keeps
	// the JSON contract explicit at the CLI boundary.
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BALDE_EXPERIMENTAL", "1")
	setupInitBudget(t)

	pluginPath := filepath.Join(dir, ".balde", "plugins", "src", "hello", "bin", "plugin")
	script := "#!/bin/sh\nwhile IFS= read -r line; do case \"$line\" in *initialize*) echo '{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"name\":\"hello\",\"protocol\":1}}';; *command/execute*) echo '{\"jsonrpc\":\"2.0\",\"id\":2,\"result\":{\"text\":\"ignored\",\"json\":{\"ok\":true}}}'; exit;; esac; done\n"
	if err := os.MkdirAll(filepath.Dir(pluginPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pluginPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(pluginPath)
	sum := sha256.Sum256(data)
	lock := `{"plugins":[{"name":"hello","version":"0.1.0","source":"fixture","sha":"x","artifact_hash":"sha256:` + hex.EncodeToString(sum[:]) + `","run":"bin/plugin","capabilities":[{"type":"command","name":"hello"}],"permissions":{}}]}`
	if err := os.WriteFile(filepath.Join(dir, ".balde", "plugins", "lock.json"), []byte(lock), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"hello", "--json"})
	cmd.SetOut(&buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("installed plugin --json error = %v", err)
	}
	if strings.TrimSpace(buf.String()) != `{"ok":true}` {
		t.Errorf("JSON output = %q, want plugin JSON", buf.String())
	}
}

func TestInstalledPluginCommandRefusedWithoutOptIn(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BALDE_EXPERIMENTAL", "")
	setupInitBudget(t)

	pluginPath := filepath.Join(dir, ".balde", "plugins", "src", "hello", "bin", "plugin")
	if err := os.MkdirAll(filepath.Dir(pluginPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pluginPath, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(pluginPath)
	sum := sha256.Sum256(data)
	lock := `{"plugins":[{"name":"hello","version":"0.1.0","source":"fixture","sha":"x","artifact_hash":"sha256:` + hex.EncodeToString(sum[:]) + `","run":"bin/plugin","capabilities":[{"type":"command","name":"hello"}],"permissions":{}}]}`
	if err := os.WriteFile(filepath.Join(dir, ".balde", "plugins", "lock.json"), []byte(lock), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"hello"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("installed plugin command without opt-in = nil error, want refusal")
	}
	if !strings.Contains(err.Error(), "experimental") {
		t.Errorf("error = %q, want experimental opt-in guidance", err)
	}
}
