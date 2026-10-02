package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// LockEntry records one installed plugin: where it came from, the exact
// commit it is pinned to, and the artifact hash verified before every run.
type LockEntry struct {
	Name         string       `json:"name"`
	Version      string       `json:"version"`
	Source       string       `json:"source"`
	Path         string       `json:"path,omitempty"`
	SHA          string       `json:"sha"`
	ArtifactHash string       `json:"artifact_hash"`
	Capabilities []Capability `json:"capabilities"`
	Permissions  Permissions  `json:"permissions"`
	Skill        string       `json:"skill,omitempty"`
	Run          string       `json:"run"`
}

// HasSkill reports whether the installed plugin ships an agent skill.
func (e LockEntry) HasSkill() bool {
	return e.Skill != ""
}

// Lockfile is the project-local plugin registry at
// <projectDir>/.balde/plugins/lock.json.
type Lockfile struct {
	Plugins []LockEntry `json:"plugins"`
}

// NewLockfile returns an empty lockfile.
func NewLockfile() *Lockfile {
	return &Lockfile{Plugins: []LockEntry{}}
}

// LockPath returns the lockfile location for a project directory.
func LockPath(projectDir string) string {
	return filepath.Join(projectDir, ".balde", "plugins", "lock.json")
}

// ReadLockfile loads the project's lockfile. A missing lockfile is an empty
// one, not an error.
func ReadLockfile(projectDir string) (Lockfile, error) {
	data, err := os.ReadFile(LockPath(projectDir))
	if os.IsNotExist(err) {
		return *NewLockfile(), nil
	}
	if err != nil {
		return Lockfile{}, fmt.Errorf("read lockfile: %w", err)
	}
	var lock Lockfile
	if err := json.Unmarshal(data, &lock); err != nil {
		return Lockfile{}, fmt.Errorf("parse lockfile: %w", err)
	}
	if lock.Plugins == nil {
		lock.Plugins = []LockEntry{}
	}
	return lock, nil
}

// Save writes the lockfile, creating the plugins directory if needed.
func (l *Lockfile) Save(projectDir string) error {
	path := LockPath(projectDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create plugins dir: %w", err)
	}
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return fmt.Errorf("encode lockfile: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write lockfile: %w", err)
	}
	return nil
}

// Add appends an entry, rejecting duplicate names.
func (l *Lockfile) Add(entry LockEntry) error {
	for _, existing := range l.Plugins {
		if existing.Name == entry.Name {
			return errors.New("lockfile: plugin " + entry.Name + " already installed")
		}
	}
	l.Plugins = append(l.Plugins, entry)
	return nil
}

// Get returns the lockfile entry for name.
func (l *Lockfile) Get(name string) (LockEntry, bool) {
	for _, entry := range l.Plugins {
		if entry.Name == name {
			return entry, true
		}
	}
	return LockEntry{}, false
}
