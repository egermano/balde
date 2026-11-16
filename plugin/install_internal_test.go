package plugin

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunBuildSurvivesBrokenVCSState verifies plugin builds do not depend on
// the VCS state of whatever repository encloses the staged sources. A `.git`
// file pointing at a missing gitdir makes every git command here exit 128 —
// the same failure `go build` reports as "error obtaining VCS status" when
// VCS stamping is inherited from the host environment.
func TestRunBuildSurvivesBrokenVCSState(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	dir := t.TempDir()
	writeInternal(t, filepath.Join(dir, "go.mod"), "module example.com/broken\n\ngo 1.26\n")
	writeInternal(t, filepath.Join(dir, "main.go"), "package main\n\nfunc main() {}\n")
	// A .git file with an unreachable gitdir forces git to exit 128 here.
	writeInternal(t, filepath.Join(dir, ".git"), "gitdir: /nonexistent/balde-test-gitdir\n")

	if err := runBuild(dir, "go build -o bin/plugin ."); err != nil {
		t.Fatalf("runBuild inside broken VCS repo: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "bin", "plugin")); err != nil {
		t.Fatalf("build artifact missing: %v", err)
	}
}

func TestWithBuildVCSDisabled(t *testing.T) {
	tests := []struct {
		name string
		env  []string
		want []string
	}{
		{
			name: "no GOFLAGS set",
			env:  []string{"PATH=/usr/bin", "HOME=/home/u"},
			want: []string{"PATH=/usr/bin", "HOME=/home/u", "GOFLAGS=-buildvcs=false"},
		},
		{
			name: "existing GOFLAGS appended",
			env:  []string{"GOFLAGS=-mod=readonly"},
			want: []string{"GOFLAGS=-mod=readonly -buildvcs=false"},
		},
		{
			name: "explicit -buildvcs choice respected",
			env:  []string{"GOFLAGS=-buildvcs=true"},
			want: []string{"GOFLAGS=-buildvcs=true"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := withBuildVCSDisabled(tt.env)
			if len(got) != len(tt.want) {
				t.Fatalf("env = %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("env[%d] = %q, want %q (full: %v)", i, got[i], tt.want[i], got)
				}
			}
		})
	}
	if !strings.Contains(strings.Join(withBuildVCSDisabled(nil), " "), "-buildvcs=false") {
		t.Error("nil env must gain GOFLAGS=-buildvcs=false")
	}
}

func writeInternal(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
