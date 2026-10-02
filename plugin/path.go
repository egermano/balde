package plugin

import "path/filepath"

// SourceDir returns the installed source directory for a plugin.
func SourceDir(projectDir, name string) string {
	if absolute, err := filepath.Abs(projectDir); err == nil {
		projectDir = absolute
	}
	return filepath.Join(projectDir, ".balde", "plugins", "src", name)
}
