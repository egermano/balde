// Package plugin implements the balde plugin system: manifest parsing,
// installation, the lockfile and the out-of-process plugin runtime.
//
// Plugins are external processes — core stays pure and untouched; the app
// service is the only data path plugins can reach, via mediated host calls.
package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// CapabilityType is the kind of extension point a plugin registers.
type CapabilityType string

const (
	CapCommand   CapabilityType = "command"
	CapImporter  CapabilityType = "importer"
	CapReporter  CapabilityType = "reporter"
	CapConnector CapabilityType = "connector"
)

// Capability is one extension point declared by a plugin.
type Capability struct {
	Type        CapabilityType `json:"type"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
}

// Permissions declare what data the plugin may read and write through host
// functions. Anything not declared is refused (E_PERM).
type Permissions struct {
	Read  []string `json:"read,omitempty"`
	Write []string `json:"write,omitempty"`
}

// Entrypoint says how to produce and run the plugin executable, relative to
// the plugin source directory.
type Entrypoint struct {
	Build string `json:"build,omitempty"`
	Run   string `json:"run"`
}

// Manifest is the plugin declaration parsed from balde-plugin.json.
type Manifest struct {
	Name         string       `json:"name"`
	Version      string       `json:"version"`
	Protocol     int          `json:"protocol"`
	Title        string       `json:"title,omitempty"`
	Capabilities []Capability `json:"capabilities"`
	Permissions  Permissions  `json:"permissions"`
	Entrypoint   Entrypoint   `json:"entrypoint"`
	Skill        string       `json:"skill,omitempty"`
}

// ParseManifest parses and validates a balde-plugin.json document.
func ParseManifest(data []byte) (Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, err
	}
	if err := m.validate(); err != nil {
		return Manifest{}, err
	}
	return m, nil
}

var validName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

var (
	validReadScopes  = map[string]bool{"accounts": true, "buckets": true, "transactions": true}
	validWriteScopes = map[string]bool{"transactions": true, "allocate": true, "buckets": true}
	validCapTypes    = map[CapabilityType]bool{
		CapCommand:   true,
		CapImporter:  true,
		CapReporter:  true,
		CapConnector: true,
	}
)

func (m Manifest) validate() error {
	if m.Name == "" {
		return errors.New("manifest: name is required")
	}
	if !validName.MatchString(m.Name) {
		return fmt.Errorf("manifest: name %q must be lowercase letters, digits and dashes", m.Name)
	}
	if m.Version == "" {
		return errors.New("manifest: version is required")
	}
	if m.Protocol != 1 {
		return fmt.Errorf("manifest: unsupported protocol %d (want 1)", m.Protocol)
	}
	if len(m.Capabilities) == 0 {
		return errors.New("manifest: at least one capability is required")
	}
	for i, cap := range m.Capabilities {
		if !validCapTypes[cap.Type] {
			return fmt.Errorf("manifest: capability %d has unknown type %q", i, cap.Type)
		}
		if cap.Type == CapCommand && cap.Name == "" {
			return fmt.Errorf("manifest: capability %d of type command requires a name", i)
		}
	}
	for _, scope := range m.Permissions.Read {
		if !validReadScopes[scope] {
			return fmt.Errorf("manifest: permission read scope %q is not allowed", scope)
		}
	}
	for _, scope := range m.Permissions.Write {
		if !validWriteScopes[scope] {
			return fmt.Errorf("manifest: permission write scope %q is not allowed", scope)
		}
	}
	if m.Entrypoint.Run == "" {
		return errors.New("manifest: entrypoint run is required")
	}
	if err := validateRunPath(m.Entrypoint.Run); err != nil {
		return err
	}
	return nil
}

// validateRunPath ensures the entrypoint stays inside the plugin directory:
// relative, no ".." segments. Absolute paths and traversals are rejected.
func validateRunPath(run string) error {
	if filepath.IsAbs(run) {
		return fmt.Errorf("manifest: entrypoint run %q must be a relative path", run)
	}
	for _, part := range strings.Split(run, string(filepath.Separator)) {
		if part == ".." {
			return fmt.Errorf("manifest: entrypoint run %q must not escape the plugin directory", run)
		}
	}
	return nil
}
