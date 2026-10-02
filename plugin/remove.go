package plugin

import (
	"fmt"
	"os"
	"path/filepath"
)

// Remove uninstalls a plugin: deletes its source directory and its lockfile
// entry.
func Remove(projectDir, name string) error {
	lock, err := ReadLockfile(projectDir)
	if err != nil {
		return fmt.Errorf("remove: %w", err)
	}
	if _, ok := lock.Get(name); !ok {
		return fmt.Errorf("remove: plugin %s is not installed", name)
	}

	srcDir := filepath.Join(projectDir, ".balde", "plugins", "src", name)
	if err := os.RemoveAll(srcDir); err != nil {
		return fmt.Errorf("remove: %w", err)
	}

	remaining := make([]LockEntry, 0, len(lock.Plugins))
	for _, entry := range lock.Plugins {
		if entry.Name != name {
			remaining = append(remaining, entry)
		}
	}
	lock.Plugins = remaining
	if err := lock.Save(projectDir); err != nil {
		return fmt.Errorf("remove: %w", err)
	}

	if err := removeSkill(projectDir, name); err != nil {
		return fmt.Errorf("remove: %w", err)
	}
	return nil
}
