package plugin_test

import (
	"strings"
	"testing"

	"github.com/egermano/balde/plugin"
)

// manifestWith returns a copy of the valid manifest with JSON substitutions
// applied, so each case mutates exactly one field.
func manifestWith(t *testing.T, replacements map[string]string) []byte {
	t.Helper()
	doc := validManifest
	for from, to := range replacements {
		if !strings.Contains(doc, from) {
			t.Fatalf("fixture does not contain %q", from)
		}
		doc = strings.Replace(doc, from, to, 1)
	}
	return []byte(doc)
}

func TestParseManifestInvalid(t *testing.T) {
	tests := []struct {
		name string
		doc  []byte
		want string // substring of the expected error
	}{
		{
			name: "missing name",
			doc:  manifestWith(t, map[string]string{`"name": "vacation",`: ``}),
			want: "name",
		},
		{
			name: "uppercase name",
			doc:  manifestWith(t, map[string]string{`"vacation"`: `"Vacation"`}),
			want: "name",
		},
		{
			name: "name with spaces",
			doc:  manifestWith(t, map[string]string{`"name": "vacation",`: `"name": "my plugin",`}),
			want: "name",
		},
		{
			name: "missing version",
			doc:  manifestWith(t, map[string]string{`"version": "0.1.0",`: ``}),
			want: "version",
		},
		{
			name: "unsupported protocol",
			doc:  manifestWith(t, map[string]string{`"protocol": 1,`: `"protocol": 2,`}),
			want: "protocol",
		},
		{
			name: "no capabilities",
			doc: manifestWith(t, map[string]string{
				`[{"type": "command", "name": "vacation", "description": "Plan vacation savings"}]`: `[]`,
			}),
			want: "capabilit",
		},
		{
			name: "unknown capability type",
			doc: manifestWith(t, map[string]string{
				`{"type": "command", "name": "vacation", "description": "Plan vacation savings"}`: `{"type": "wizard"}`,
			}),
			want: "capabilit",
		},
		{
			name: "command capability without name",
			doc: manifestWith(t, map[string]string{
				`{"type": "command", "name": "vacation", "description": "Plan vacation savings"}`: `{"type": "command"}`,
			}),
			want: "name",
		},
		{
			name: "unknown read scope",
			doc:  manifestWith(t, map[string]string{`["accounts", "buckets", "transactions"]`: `["accounts", "secrets"]`}),
			want: "permission",
		},
		{
			name: "unknown write scope",
			doc:  manifestWith(t, map[string]string{`["transactions", "allocate"]`: `["everything"]`}),
			want: "permission",
		},
		{
			name: "missing entrypoint run",
			doc:  manifestWith(t, map[string]string{`{"build": "go build -o bin/vacation .", "run": "bin/vacation"}`: `{"build": "make"}`}),
			want: "entrypoint",
		},
		{
			name: "entrypoint run escapes plugin dir",
			doc:  manifestWith(t, map[string]string{`"run": "bin/vacation"`: `"run": "../evil"`}),
			want: "entrypoint",
		},
		{
			name: "entrypoint run absolute path",
			doc:  manifestWith(t, map[string]string{`"run": "bin/vacation"`: `"run": "/bin/sh"`}),
			want: "entrypoint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := plugin.ParseManifest(tt.doc)
			if err == nil {
				t.Fatalf("ParseManifest() = nil error, want error containing %q", tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("ParseManifest() error = %q, want it to contain %q", err, tt.want)
			}
		})
	}
}

func TestParseManifestValidScopes(t *testing.T) {
	// every documented read/write scope must be accepted
	doc := manifestWith(t, map[string]string{
		`["accounts", "buckets", "transactions"]`: `["accounts", "buckets", "transactions"]`,
		`["transactions", "allocate"]`:            `["transactions", "allocate", "buckets"]`,
	})
	if _, err := plugin.ParseManifest(doc); err != nil {
		t.Fatalf("ParseManifest() error = %v, want nil", err)
	}
}
