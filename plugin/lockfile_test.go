package plugin_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/egermano/balde/plugin"
)

func TestLockfileMissingFileIsEmpty(t *testing.T) {
	project := t.TempDir()

	lock, err := plugin.ReadLockfile(project)
	if err != nil {
		t.Fatalf("ReadLockfile() error = %v, want nil", err)
	}
	if len(lock.Plugins) != 0 {
		t.Errorf("Plugins = %d, want 0", len(lock.Plugins))
	}
}

func TestLockfileSaveAndReadRoundtrip(t *testing.T) {
	project := t.TempDir()

	entry := plugin.LockEntry{
		Name:         "vacation",
		Version:      "0.1.0",
		Source:       "github.com/user/balde-plugin-vacation",
		SHA:          "abc123",
		ArtifactHash: "sha256:deadbeef",
		Capabilities: []plugin.Capability{{Type: plugin.CapCommand, Name: "vacation"}},
		Permissions:  plugin.Permissions{Read: []string{"accounts", "buckets"}},
	}

	lock := plugin.NewLockfile()
	if err := lock.Add(entry); err != nil {
		t.Fatalf("Add() error = %v, want nil", err)
	}
	if err := lock.Save(project); err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}

	wantPath := filepath.Join(project, ".balde", "plugins", "lock.json")
	if _, err := os.Stat(wantPath); err != nil {
		t.Fatalf("lockfile not written to %s: %v", wantPath, err)
	}

	reread, err := plugin.ReadLockfile(project)
	if err != nil {
		t.Fatalf("ReadLockfile() error = %v, want nil", err)
	}
	got, ok := reread.Get("vacation")
	if !ok {
		t.Fatalf("Get(vacation) not found in %v", reread.Plugins)
	}
	if !reflect.DeepEqual(got, entry) {
		t.Errorf("roundtrip = %+v, want %+v", got, entry)
	}
}

func TestLockfileAddDuplicateNameRejected(t *testing.T) {
	lock := plugin.NewLockfile()
	entry := plugin.LockEntry{Name: "vacation", Version: "0.1.0", Source: "s"}

	if err := lock.Add(entry); err != nil {
		t.Fatalf("first Add() error = %v, want nil", err)
	}
	if err := lock.Add(entry); err == nil {
		t.Error("second Add() = nil error, want duplicate error")
	}
}
