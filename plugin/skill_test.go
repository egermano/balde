package plugin_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/egermano/balde/plugin"
)

const validSkill = `---
name: balde-vacation
description: Plan and track vacation savings with balde. Use whenever the user mentions a trip, vacation, holiday savings or travel budget.
---

# Vacation planner skill

Instructions here.
`

// fixtureRepoWithSkill creates a git repo whose plugin (at subpath) declares
// and ships a skill directory.
func fixtureRepoWithSkill(t *testing.T, skillManifest, skillContent string) string {
	t.Helper()

	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "plugins", "vacation")
	if err := os.MkdirAll(filepath.Join(pluginDir, "skill"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(pluginDir, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(pluginDir, "balde-plugin.json"), skillManifest)
	writeFile(t, filepath.Join(pluginDir, "bin", "plugin"), "#!/bin/sh\ntrue\n")
	if skillContent != "" {
		writeFile(t, filepath.Join(pluginDir, "skill", "SKILL.md"), skillContent)
	}
	git(t, dir, "init", "-q")
	git(t, dir, "add", "-A")
	git(t, dir, "-c", "user.name=test", "-c", "user.email=test@test", "commit", "-q", "-m", "init")
	return dir
}

const manifestWithSkill = `{
  "name": "vacation",
  "version": "0.1.0",
  "protocol": 1,
  "capabilities": [{"type": "command", "name": "vacation", "description": "plan"}],
  "permissions": {"read": ["buckets"]},
  "entrypoint": {"run": "bin/plugin"},
  "skill": "skill"
}`

func TestInstallInstallsSkill(t *testing.T) {
	project := t.TempDir()
	repo := fixtureRepoWithSkill(t, manifestWithSkill, validSkill)

	entry, err := plugin.Install(project, repo, "plugins/vacation")
	if err != nil {
		t.Fatalf("Install() error = %v, want nil", err)
	}
	if !entry.HasSkill() {
		t.Error("entry.HasSkill() = false, want true")
	}

	skillDir := filepath.Join(project, ".agents", "skills", "balde-vacation")
	data, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("skill not installed: %v", err)
	}
	if string(data) != validSkill {
		t.Errorf("SKILL.md content = %q, want original", data)
	}

	// ownership marker must exist
	if _, err := os.Stat(filepath.Join(skillDir, ".balde-plugin-skill")); err != nil {
		t.Errorf("skill marker missing: %v", err)
	}
}

func TestInstallSkillRefusesToClobberForeignSkill(t *testing.T) {
	project := t.TempDir()
	repo := fixtureRepoWithSkill(t, manifestWithSkill, validSkill)

	foreign := filepath.Join(project, ".agents", "skills", "balde-vacation")
	if err := os.MkdirAll(foreign, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(foreign, "SKILL.md"), "user-authored skill")

	if _, err := plugin.Install(project, repo, "plugins/vacation"); err == nil {
		t.Error("Install() = nil error, want refusal over existing foreign skill")
	}
	data, _ := os.ReadFile(filepath.Join(foreign, "SKILL.md"))
	if string(data) != "user-authored skill" {
		t.Error("foreign skill was overwritten")
	}
}

func TestInstallSkillReinstallOverOwnMarker(t *testing.T) {
	project := t.TempDir()
	repo := fixtureRepoWithSkill(t, manifestWithSkill, validSkill)

	if _, err := plugin.Install(project, repo, "plugins/vacation"); err != nil {
		t.Fatalf("first Install() error = %v", err)
	}
	// remove plugin entry but keep skill, then reinstall the same plugin
	if _, err := plugin.Install(project, repo, "plugins/vacation"); err == nil {
		t.Log("second install rejected as duplicate (acceptable)")
	}
	if err := plugin.Remove(project, "vacation"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if _, err := plugin.Install(project, repo, "plugins/vacation"); err != nil {
		t.Fatalf("reinstall after remove error = %v, want nil", err)
	}
}

func TestInstallSkillInvalid(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(t *testing.T) string
	}{
		{
			name: "declared skill directory missing",
			prepare: func(t *testing.T) string {
				return fixtureRepoWithSkill(t, manifestWithSkill, "")
			},
		},
		{
			name: "skill file without frontmatter name",
			prepare: func(t *testing.T) string {
				return fixtureRepoWithSkill(t, manifestWithSkill, "---\ndescription: something\n---\nbody\n")
			},
		},
		{
			name: "skill file with wrong frontmatter name",
			prepare: func(t *testing.T) string {
				return fixtureRepoWithSkill(t, manifestWithSkill, "---\nname: other-name\ndescription: d\n---\nbody\n")
			},
		},
		{
			name: "skill file without description",
			prepare: func(t *testing.T) string {
				return fixtureRepoWithSkill(t, manifestWithSkill, "---\nname: balde-vacation\n---\nbody\n")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := t.TempDir()
			repo := tt.prepare(t)

			if _, err := plugin.Install(project, repo, "plugins/vacation"); err == nil {
				t.Error("Install() = nil error, want skill validation error")
			}
			if _, err := os.Stat(filepath.Join(project, ".agents", "skills", "balde-vacation")); !os.IsNotExist(err) {
				t.Error("skill dir should not exist after failed install")
			}
		})
	}
}

func TestRemoveAlsoRemovesSkill(t *testing.T) {
	project := t.TempDir()
	repo := fixtureRepoWithSkill(t, manifestWithSkill, validSkill)

	if _, err := plugin.Install(project, repo, "plugins/vacation"); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if err := plugin.Remove(project, "vacation"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	skillDir := filepath.Join(project, ".agents", "skills", "balde-vacation")
	if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
		t.Errorf("skill dir still present after remove: %v", err)
	}
}

func TestRemoveKeepsForeignSkillWithSameName(t *testing.T) {
	project := t.TempDir()

	// simulate: lockfile entry exists (hand-written) but skill dir is foreign
	pluginsDir := filepath.Join(project, ".balde", "plugins")
	if err := os.MkdirAll(filepath.Join(pluginsDir, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	lock := plugin.NewLockfile()
	if err := lock.Add(plugin.LockEntry{
		Name: "vacation", Version: "0.1.0", Source: "x",
		Capabilities: []plugin.Capability{{Type: plugin.CapCommand, Name: "vacation"}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := lock.Save(project); err != nil {
		t.Fatal(err)
	}

	foreign := filepath.Join(project, ".agents", "skills", "balde-vacation")
	if err := os.MkdirAll(foreign, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(foreign, "SKILL.md"), "user-authored skill")

	if err := plugin.Remove(project, "vacation"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(foreign, "SKILL.md")); err != nil {
		t.Errorf("foreign skill deleted by remove: %v", err)
	}
}
