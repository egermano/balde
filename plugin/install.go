package plugin

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Install clones the plugin source (a git repository path or URL) into the
// project, validates its manifest, runs the optional build step, verifies
// the entrypoint artifact, and records everything in the lockfile.
//
// Installing is the explicit trust grant: nothing runs at install time
// except the declared build command.
func Install(projectDir, source string) (LockEntry, error) {
	pluginsDir := filepath.Join(projectDir, ".balde", "plugins")

	tmp, err := os.MkdirTemp("", "balde-install-")
	if err != nil {
		return LockEntry{}, fmt.Errorf("install: %w", err)
	}
	defer os.RemoveAll(tmp)

	src := filepath.Join(tmp, "src")
	if out, err := gitRun(tmp, "clone", "--quiet", source, src); err != nil {
		return LockEntry{}, fmt.Errorf("install: clone: %w\n%s", err, out)
	}

	sha, err := gitRun(src, "rev-parse", "HEAD")
	if err != nil {
		return LockEntry{}, fmt.Errorf("install: resolve commit: %w", err)
	}

	data, err := os.ReadFile(filepath.Join(src, "balde-plugin.json"))
	if err != nil {
		return LockEntry{}, fmt.Errorf("install: read manifest: %w", err)
	}
	m, err := ParseManifest(data)
	if err != nil {
		return LockEntry{}, fmt.Errorf("install: %w", err)
	}

	finalDir := filepath.Join(pluginsDir, "src", m.Name)
	if _, err := os.Stat(finalDir); err == nil {
		return LockEntry{}, fmt.Errorf("install: plugin %s already installed", m.Name)
	}
	if err := os.MkdirAll(filepath.Dir(finalDir), 0o755); err != nil {
		return LockEntry{}, fmt.Errorf("install: %w", err)
	}
	if err := os.Rename(src, finalDir); err != nil {
		return LockEntry{}, fmt.Errorf("install: %w", err)
	}

	if m.Entrypoint.Build != "" {
		cmd := exec.Command("sh", "-c", m.Entrypoint.Build)
		cmd.Dir = finalDir
		if out, err := cmd.CombinedOutput(); err != nil {
			os.RemoveAll(finalDir)
			return LockEntry{}, fmt.Errorf("install: build: %w\n%s", err, out)
		}
	}

	artifact := filepath.Join(finalDir, m.Entrypoint.Run)
	file, err := os.Open(artifact)
	if err != nil {
		os.RemoveAll(finalDir)
		return LockEntry{}, fmt.Errorf("install: entrypoint %s not found: %w", m.Entrypoint.Run, err)
	}
	defer file.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		os.RemoveAll(finalDir)
		return LockEntry{}, fmt.Errorf("install: hash artifact: %w", err)
	}

	entry := LockEntry{
		Name:         m.Name,
		Version:      m.Version,
		Source:       source,
		SHA:          strings.TrimSpace(sha),
		ArtifactHash: "sha256:" + hex.EncodeToString(hasher.Sum(nil)),
		Capabilities: m.Capabilities,
		Permissions:  m.Permissions,
	}

	lock, err := ReadLockfile(projectDir)
	if err != nil {
		os.RemoveAll(finalDir)
		return LockEntry{}, fmt.Errorf("install: %w", err)
	}
	if err := lock.Add(entry); err != nil {
		os.RemoveAll(finalDir)
		return LockEntry{}, fmt.Errorf("install: %w", err)
	}
	if err := lock.Save(projectDir); err != nil {
		os.RemoveAll(finalDir)
		return LockEntry{}, fmt.Errorf("install: %w", err)
	}
	return entry, nil
}

func gitRun(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}
