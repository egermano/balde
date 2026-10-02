package plugin_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/egermano/balde/plugin"
)

// fixtureRepo creates a local git repository containing a valid plugin
// manifest, an executable entrypoint, and returns its path.
func fixtureRepo(t *testing.T, name string) string {
	t.Helper()

	dir := t.TempDir()
	manifest := `{
  "name": "` + name + `",
  "version": "0.1.0",
  "protocol": 1,
  "capabilities": [{"type": "command", "name": "` + name + `", "description": "test plugin"}],
  "permissions": {"read": ["accounts", "buckets"]},
  "entrypoint": {"run": "bin/plugin"}
}`
	if err := os.MkdirAll(filepath.Join(dir, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "balde-plugin.json"), manifest)
	writeFile(t, filepath.Join(dir, "bin", "plugin"), "#!/bin/sh\necho hello\n")
	git(t, dir, "init", "-q")
	git(t, dir, "add", "-A")
	git(t, dir, "-c", "user.name=test", "-c", "user.email=test@test", "commit", "-q", "-m", "init")
	return dir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

// fixtureRepoSubdir creates a git repository whose plugin lives at subpath
// inside the repo (non-root plugin folder) and returns the repo path.
func fixtureRepoSubdir(t *testing.T, name, subpath string) string {
	t.Helper()

	dir := t.TempDir()
	pluginDir := filepath.Join(dir, subpath)
	manifest := `{
  "name": "` + name + `",
  "version": "0.1.0",
  "protocol": 1,
  "capabilities": [{"type": "command", "name": "` + name + `", "description": "test plugin"}],
  "permissions": {"read": ["buckets"]},
  "entrypoint": {"run": "bin/plugin"}
}`
	if err := os.MkdirAll(filepath.Join(pluginDir, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(pluginDir, "balde-plugin.json"), manifest)
	writeFile(t, filepath.Join(pluginDir, "bin", "plugin"), "#!/bin/sh\necho hello\n")
	git(t, dir, "init", "-q")
	git(t, dir, "add", "-A")
	git(t, dir, "-c", "user.name=test", "-c", "user.email=test@test", "commit", "-q", "-m", "init")
	return dir
}

func TestInstallFromSubdirectory(t *testing.T) {
	project := t.TempDir()
	repo := fixtureRepoSubdir(t, "vacation", "plugins/vacation")

	entry, err := plugin.Install(project, repo, "plugins/vacation")
	if err != nil {
		t.Fatalf("Install() error = %v, want nil", err)
	}

	if entry.Path != "plugins/vacation" {
		t.Errorf("Path = %q, want plugins/vacation", entry.Path)
	}

	// only the subdirectory content is installed as plugin source
	artifact := filepath.Join(project, ".balde", "plugins", "src", "vacation", "bin", "plugin")
	if _, err := os.Stat(artifact); err != nil {
		t.Errorf("artifact not installed at %s: %v", artifact, err)
	}
	if _, err := os.Stat(filepath.Join(project, ".balde", "plugins", "src", "vacation", "balde-plugin.json")); err != nil {
		t.Errorf("manifest not installed: %v", err)
	}
}

func TestInstallSubdirectoryMissingManifest(t *testing.T) {
	project := t.TempDir()
	repo := fixtureRepoSubdir(t, "vacation", "plugins/vacation")

	if _, err := plugin.Install(project, repo, "plugins/elsewhere"); err == nil {
		t.Error("Install(missing subpath) = nil error, want error")
	}
}

func TestInstallSubdirectoryCannotEscapeRepository(t *testing.T) {
	project := t.TempDir()
	repo := fixtureRepoSubdir(t, "vacation", "plugins/vacation")

	if _, err := plugin.Install(project, repo, "../outside"); err == nil {
		t.Error("Install(escaping subpath) = nil error, want error")
	} else if !strings.Contains(err.Error(), "path") {
		t.Errorf("error = %q, want path validation error", err)
	}
}

func TestInstallLeavesNoStagingDirectories(t *testing.T) {
	project := t.TempDir()
	repo := fixtureRepo(t, "vacation")

	if _, err := plugin.Install(project, repo, ""); err != nil {
		t.Fatalf("Install() error = %v, want nil", err)
	}
	if _, err := plugin.Install(project, t.TempDir(), ""); err == nil {
		t.Fatal("expected duplicate-install error")
	}

	pluginsDir := filepath.Join(project, ".balde", "plugins")
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".install-") {
			t.Errorf("staging directory %s left behind in %s", e.Name(), pluginsDir)
		}
	}
}

func TestInstallFromLocalRepo(t *testing.T) {
	project := t.TempDir()
	repo := fixtureRepo(t, "vacation")

	entry, err := plugin.Install(project, repo, "")
	if err != nil {
		t.Fatalf("Install() error = %v, want nil", err)
	}

	if entry.Name != "vacation" {
		t.Errorf("Name = %q, want vacation", entry.Name)
	}
	if entry.Source != repo {
		t.Errorf("Source = %q, want %q", entry.Source, repo)
	}
	if len(entry.SHA) != 40 {
		t.Errorf("SHA = %q, want 40-char commit hash", entry.SHA)
	}
	if !strings.HasPrefix(entry.ArtifactHash, "sha256:") {
		t.Errorf("ArtifactHash = %q, want sha256: prefix", entry.ArtifactHash)
	}

	// plugin source moved into the project
	artifact := filepath.Join(project, ".balde", "plugins", "src", "vacation", "bin", "plugin")
	if _, err := os.Stat(artifact); err != nil {
		t.Errorf("artifact not installed at %s: %v", artifact, err)
	}

	// lockfile contains the entry
	lock, err := plugin.ReadLockfile(project)
	if err != nil {
		t.Fatalf("ReadLockfile() error = %v", err)
	}
	got, ok := lock.Get("vacation")
	if !ok {
		t.Fatalf("lockfile missing vacation entry: %+v", lock.Plugins)
	}
	if got.SHA != entry.SHA {
		t.Errorf("lockfile SHA = %q, want %q", got.SHA, entry.SHA)
	}
}

func TestInstallDuplicateRejected(t *testing.T) {
	project := t.TempDir()
	repo := fixtureRepo(t, "vacation")

	if _, err := plugin.Install(project, repo, ""); err != nil {
		t.Fatalf("first Install() error = %v, want nil", err)
	}
	if _, err := plugin.Install(project, repo, ""); err == nil {
		t.Error("second Install() = nil error, want already-installed error")
	}
}

func TestInstallInvalidManifestRejected(t *testing.T) {
	project := t.TempDir()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "balde-plugin.json"), `{"name": "bad"}`)
	git(t, dir, "init", "-q")
	git(t, dir, "add", "-A")
	git(t, dir, "-c", "user.name=test", "-c", "user.email=test@test", "commit", "-q", "-m", "init")

	if _, err := plugin.Install(project, dir, ""); err == nil {
		t.Error("Install(invalid manifest) = nil error, want error")
	}

	// nothing should be left behind
	srcDir := filepath.Join(project, ".balde", "plugins", "src")
	if entries, _ := os.ReadDir(srcDir); len(entries) != 0 {
		t.Errorf("src dir = %v, want empty after failed install", entries)
	}
}
