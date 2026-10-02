package plugin_test

import (
	"testing"

	"github.com/egermano/balde/plugin"
)

func TestResolveSource(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "github shorthand becomes https url",
			source: "github.com/user/balde-plugin-vacation",
			want:   "https://github.com/user/balde-plugin-vacation.git",
		},
		{
			name:   "https url unchanged",
			source: "https://github.com/user/repo.git",
			want:   "https://github.com/user/repo.git",
		},
		{
			name:   "ssh remote unchanged",
			source: "git@github.com:user/repo.git",
			want:   "git@github.com:user/repo.git",
		},
		{
			name:   "local path unchanged",
			source: "/tmp/fixture-repo",
			want:   "/tmp/fixture-repo",
		},
		{
			name:   "relative path unchanged",
			source: "../some/repo",
			want:   "../some/repo",
		},
		{
			name:   "other host shorthand unchanged",
			source: "gitlab.com/user/repo",
			want:   "gitlab.com/user/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := plugin.ResolveSource(tt.source); got != tt.want {
				t.Errorf("ResolveSource(%q) = %q, want %q", tt.source, got, tt.want)
			}
		})
	}
}
