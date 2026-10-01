package plugin_test

import (
	"testing"

	"github.com/egermano/balde/plugin"
)

const validManifest = `{
  "name": "vacation",
  "version": "0.1.0",
  "protocol": 1,
  "title": "Vacation planner",
  "capabilities": [
    {"type": "command", "name": "vacation", "description": "Plan vacation savings"}
  ],
  "permissions": {
    "read": ["accounts", "buckets", "transactions"],
    "write": ["transactions", "allocate"]
  },
  "entrypoint": {"build": "go build -o bin/vacation .", "run": "bin/vacation"}
}`

func TestParseManifestValid(t *testing.T) {
	m, err := plugin.ParseManifest([]byte(validManifest))
	if err != nil {
		t.Fatalf("ParseManifest() error = %v, want nil", err)
	}

	if m.Name != "vacation" {
		t.Errorf("Name = %q, want vacation", m.Name)
	}
	if m.Version != "0.1.0" {
		t.Errorf("Version = %q, want 0.1.0", m.Version)
	}
	if m.Protocol != 1 {
		t.Errorf("Protocol = %d, want 1", m.Protocol)
	}
	if len(m.Capabilities) != 1 || m.Capabilities[0].Type != plugin.CapCommand {
		t.Errorf("Capabilities = %+v, want one command capability", m.Capabilities)
	}
	if m.Capabilities[0].Name != "vacation" {
		t.Errorf("command capability name = %q, want vacation", m.Capabilities[0].Name)
	}
	if len(m.Permissions.Read) != 3 || len(m.Permissions.Write) != 2 {
		t.Errorf("Permissions = %+v, want 3 read / 2 write", m.Permissions)
	}
	if m.Entrypoint.Run != "bin/vacation" {
		t.Errorf("Entrypoint.Run = %q, want bin/vacation", m.Entrypoint.Run)
	}
	if m.Entrypoint.Build == "" {
		t.Error("Entrypoint.Build = empty, want the build command")
	}
}
