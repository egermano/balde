// Package plugin implements the balde plugin system: manifest parsing,
// installation, the lockfile and the out-of-process plugin runtime.
//
// Plugins are external processes — core stays pure and untouched; the app
// service is the only data path plugins can reach, via mediated host calls.
package plugin

import "encoding/json"

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
}

// ParseManifest parses and validates a balde-plugin.json document.
func ParseManifest(data []byte) (Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, err
	}
	return m, nil
}
