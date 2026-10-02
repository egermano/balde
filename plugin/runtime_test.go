package plugin_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/egermano/balde/app"
	"github.com/egermano/balde/plugin"
	"github.com/egermano/balde/store"
)

var (
	fakeOnce sync.Once
	fakePath string
	fakeErr  error
)

// fakePlugin builds (once) the Go fixture plugin used by runtime tests.
func fakePlugin(t *testing.T) string {
	t.Helper()
	fakeOnce.Do(func() {
		dir, err := os.MkdirTemp("", "balde-fakeplugin-*")
		if err != nil {
			fakeErr = err
			return
		}
		fakePath = filepath.Join(dir, "fakeplugin")
		out, err := exec.Command("go", "build", "-o", fakePath, "./testdata/fakeplugin").CombinedOutput()
		if err != nil {
			fakeErr = fmt.Errorf("build fake plugin: %v\n%s", err, out)
		}
	})
	if fakeErr != nil {
		t.Fatal(fakeErr)
	}
	return fakePath
}

func fileHash(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func copyFile(t *testing.T, src, dest string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dest, data, 0o755); err != nil {
		t.Fatal(err)
	}
}

// installFakePlugin stages the fixture binary as an installed plugin and
// returns its lock entry.
func installFakePlugin(t *testing.T, project string, perms plugin.Permissions) plugin.LockEntry {
	t.Helper()

	bin := fakePlugin(t)
	srcDir := filepath.Join(project, ".balde", "plugins", "src", "fake")
	copyFile(t, bin, filepath.Join(srcDir, "bin", "fake"))

	entry := plugin.LockEntry{
		Name:         "fake",
		Version:      "0.1.0",
		SHA:          "testsha",
		Run:          "bin/fake",
		ArtifactHash: fileHash(t, filepath.Join(srcDir, "bin", "fake")),
		Capabilities: []plugin.Capability{{Type: plugin.CapCommand, Name: "fake"}},
		Permissions:  perms,
	}
	lock := plugin.NewLockfile()
	if err := lock.Add(entry); err != nil {
		t.Fatal(err)
	}
	if err := lock.Save(project); err != nil {
		t.Fatal(err)
	}
	return entry
}

func newTestRunner(t *testing.T, entry plugin.LockEntry, project string) *plugin.Runner {
	t.Helper()

	db, err := store.NewSQLiteStore(filepath.Join(project, "balde.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	return &plugin.Runner{
		Entry:  entry,
		SrcDir: filepath.Join(project, ".balde", "plugins", "src", entry.Name),
		App:    app.New("default", db),
		Meta: plugin.BudgetMeta{
			CurrencySymbol:     "$",
			DecimalSeparator:   ".",
			ThousandsSeparator: ",",
			Frequency:          "monthly",
		},
	}
}

func TestRunnerExecutesCommand(t *testing.T) {
	project := t.TempDir()
	entry := installFakePlugin(t, project, plugin.Permissions{})
	runner := newTestRunner(t, entry, project)

	result, err := runner.Run("ok", nil)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if result.Text != "hello from plugin" {
		t.Errorf("Text = %q, want hello from plugin", result.Text)
	}
}

func TestRunnerHostCallWithPermission(t *testing.T) {
	project := t.TempDir()
	entry := installFakePlugin(t, project, plugin.Permissions{Read: []string{"buckets"}})
	runner := newTestRunner(t, entry, project)

	// seed one bucket through the app service
	if _, err := runner.App.AddBucket("goals", 50000); err != nil {
		t.Fatal(err)
	}

	result, err := runner.Run("list-buckets", nil)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if !strings.Contains(result.Text, "goals") {
		t.Errorf("Text = %q, want bucket data", result.Text)
	}
}

func TestRunnerHostCallWithoutPermission(t *testing.T) {
	project := t.TempDir()
	entry := installFakePlugin(t, project, plugin.Permissions{}) // no scopes
	runner := newTestRunner(t, entry, project)

	result, err := runner.Run("list-buckets", nil)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil (E_PERM is returned to the plugin)", err)
	}
	if !strings.Contains(result.Text, "E_PERM") {
		t.Errorf("Text = %q, want E_PERM surfaced to the plugin", result.Text)
	}
}

func TestRunnerWriteHostCallMediated(t *testing.T) {
	project := t.TempDir()
	entry := installFakePlugin(t, project, plugin.Permissions{Write: []string{"buckets", "allocate"}})
	runner := newTestRunner(t, entry, project)

	if _, err := runner.Run("add-bucket", nil); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	buckets, err := runner.App.ListBuckets()
	if err != nil {
		t.Fatal(err)
	}
	if len(buckets) != 1 || buckets[0].Name != "from-plugin" {
		t.Errorf("buckets = %+v, want from-plugin created through the app service", buckets)
	}
}

func TestRunnerWriteHostCallWithoutPermissionChangesNothing(t *testing.T) {
	project := t.TempDir()
	entry := installFakePlugin(t, project, plugin.Permissions{Read: []string{"buckets"}})
	runner := newTestRunner(t, entry, project)

	if _, err := runner.Run("add-bucket", nil); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	buckets, err := runner.App.ListBuckets()
	if err != nil {
		t.Fatal(err)
	}
	if len(buckets) != 0 {
		t.Errorf("buckets = %+v, want none — undeclared write must not mutate state", buckets)
	}
}

func TestRunnerWriteHostCallPreservesBucketLimit(t *testing.T) {
	project := t.TempDir()
	entry := installFakePlugin(t, project, plugin.Permissions{Write: []string{"buckets"}})
	runner := newTestRunner(t, entry, project)

	for i := 0; i < 8; i++ {
		if _, err := runner.App.AddBucket(fmt.Sprintf("bucket-%d", i), 0); err != nil {
			t.Fatalf("seed bucket %d: %v", i, err)
		}
	}

	result, err := runner.Run("add-bucket", nil)
	if err != nil {
		t.Fatalf("Run() error = %v, want plugin-visible validation result", err)
	}
	if !strings.Contains(result.Text, "maximum of 8 buckets") {
		t.Errorf("Text = %q, want core bucket-limit validation", result.Text)
	}

	buckets, err := runner.App.ListBuckets()
	if err != nil {
		t.Fatal(err)
	}
	if len(buckets) != 8 {
		t.Errorf("bucket count = %d, want 8", len(buckets))
	}
}

func TestRunnerPluginErrorPropagates(t *testing.T) {
	project := t.TempDir()
	entry := installFakePlugin(t, project, plugin.Permissions{})
	runner := newTestRunner(t, entry, project)

	_, err := runner.Run("fail", nil)
	if err == nil {
		t.Fatal("Run(fail) = nil error, want error")
	}
	if !strings.Contains(err.Error(), "E_PLUGIN") {
		t.Errorf("error = %v, want E_PLUGIN", err)
	}
}

func TestRunnerTamperedArtifactRefused(t *testing.T) {
	project := t.TempDir()
	entry := installFakePlugin(t, project, plugin.Permissions{})

	// tamper with the installed binary after the hash was recorded
	binPath := filepath.Join(project, ".balde", "plugins", "src", "fake", "bin", "fake")
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\nevil\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	runner := newTestRunner(t, entry, project)
	_, err := runner.Run("ok", nil)
	if err == nil {
		t.Fatal("Run() with tampered artifact = nil error, want refusal")
	}
	if !strings.Contains(err.Error(), "hash mismatch") {
		t.Errorf("error = %v, want hash mismatch", err)
	}
}

func TestRunnerPassesArgsThrough(t *testing.T) {
	project := t.TempDir()
	entry := installFakePlugin(t, project, plugin.Permissions{})
	runner := newTestRunner(t, entry, project)

	result, err := runner.Run("echo-args", []string{"plan", "--date", "2027-03"})
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	want := "plan --date 2027-03"
	if result.Text != want {
		t.Errorf("Text = %q, want %q", result.Text, want)
	}
}
