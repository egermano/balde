package plugin_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/egermano/balde/plugin"
)

func TestRemoveDeletesSourceAndLockEntry(t *testing.T) {
	project := t.TempDir()
	repo := fixtureRepo(t, "vacation")

	if _, err := plugin.Install(project, repo); err != nil {
		t.Fatalf("Install() error = %v, want nil", err)
	}

	if err := plugin.Remove(project, "vacation"); err != nil {
		t.Fatalf("Remove() error = %v, want nil", err)
	}

	srcDir := filepath.Join(project, ".balde", "plugins", "src", "vacation")
	if _, err := os.Stat(srcDir); !os.IsNotExist(err) {
		t.Errorf("source dir still present: %v", err)
	}

	lock, err := plugin.ReadLockfile(project)
	if err != nil {
		t.Fatalf("ReadLockfile() error = %v, want nil", err)
	}
	if _, ok := lock.Get("vacation"); ok {
		t.Error("lockfile still contains vacation entry")
	}
}

func TestRemoveUnknownPluginErrors(t *testing.T) {
	project := t.TempDir()

	if err := plugin.Remove(project, "ghost"); err == nil {
		t.Error("Remove(unknown) = nil error, want error")
	}
}
