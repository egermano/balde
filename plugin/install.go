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

// Install clones the plugin source (a git repository path or URL), reads the
// manifest at the optional subpath (for repos hosting plugins in non-root
// folders), validates it, runs the optional build step, verifies the
// entrypoint artifact, and records everything in the lockfile.
//
// Installing is the explicit trust grant: nothing runs at install time
// except the declared build command.
func Install(projectDir, source, path string) (LockEntry, error) {
	if err := validatePluginPath(path); err != nil {
		return LockEntry{}, err
	}
	absoluteProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return LockEntry{}, fmt.Errorf("install: resolve project directory: %w", err)
	}
	projectDir = absoluteProjectDir
	pluginsDir := filepath.Join(projectDir, ".balde", "plugins")

	// Stage inside the plugins directory so the final move never crosses a
	// filesystem boundary (os.Rename fails with EXDEV across devices).
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		return LockEntry{}, fmt.Errorf("install: %w", err)
	}
	tmp, err := os.MkdirTemp(pluginsDir, ".install-")
	if err != nil {
		return LockEntry{}, fmt.Errorf("install: %w", err)
	}
	defer os.RemoveAll(tmp)

	cloneSource, ref := ParseSourceReference(source)
	cloneDir := filepath.Join(tmp, "repo")
	if out, err := gitRun(tmp, "clone", "--quiet", cloneSource, cloneDir); err != nil {
		return LockEntry{}, fmt.Errorf("install: clone: %w\n%s", err, out)
	}
	if ref != "" {
		if out, err := gitRun(cloneDir, "checkout", "--quiet", "--detach", ref); err != nil {
			return LockEntry{}, fmt.Errorf("install: checkout %q: %w\n%s", ref, err, out)
		}
	}

	src := cloneDir
	if path != "" {
		src = filepath.Join(cloneDir, filepath.FromSlash(path))
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
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		file.Close()
		os.RemoveAll(finalDir)
		return LockEntry{}, fmt.Errorf("install: hash artifact: %w", err)
	}
	file.Close()

	entry := LockEntry{
		Name:         m.Name,
		Version:      m.Version,
		Source:       source,
		Path:         path,
		SHA:          strings.TrimSpace(sha),
		ArtifactHash: "sha256:" + hex.EncodeToString(hasher.Sum(nil)),
		Capabilities: m.Capabilities,
		Permissions:  m.Permissions,
		Skill:        m.Skill,
		Run:          m.Entrypoint.Run,
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

	if _, err := installSkill(projectDir, m, finalDir); err != nil {
		// roll the whole install back: plugin + lockfile entry
		os.RemoveAll(finalDir)
		lock.Plugins = lock.Plugins[:len(lock.Plugins)-1]
		_ = lock.Save(projectDir)
		return LockEntry{}, err
	}
	return entry, nil
}

// validatePluginPath keeps --path within the cloned repository. This is
// separate from entrypoint validation because --path is supplied by the user
// at install time, not by a trusted manifest.
func validatePluginPath(path string) error {
	if path == "" {
		return nil
	}
	clean := filepath.Clean(filepath.FromSlash(path))
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("install: path %q must stay within the repository", path)
	}
	return nil
}

func gitRun(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}
