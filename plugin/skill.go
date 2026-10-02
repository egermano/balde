package plugin

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// skillMarkerFile marks a skill directory as balde-installed, so plugin
// installs never clobber user-authored skills and removes never delete them.
const skillMarkerFile = ".balde-plugin-skill"

// skillName returns the namespaced skill name for a plugin.
func skillName(pluginName string) string {
	return "balde-" + pluginName
}

// installSkill validates and copies a plugin's declared skill into the
// project's .agents/skills/ directory (the cross-tool agent skill
// convention). Refuses to overwrite a directory balde did not install.
func installSkill(projectDir string, m Manifest, pluginDir string) (string, error) {
	if m.Skill == "" {
		return "", nil
	}

	src := filepath.Join(pluginDir, filepath.FromSlash(m.Skill))
	info, err := os.Stat(src)
	if err != nil {
		return "", fmt.Errorf("install: skill %s not found: %w", m.Skill, err)
	}

	skillDir := src
	if !info.IsDir() {
		// manifest pointed at the SKILL.md file directly; its directory is
		// the skill root
		skillDir = filepath.Dir(src)
	}

	if err := validateSkill(filepath.Join(skillDir, "SKILL.md"), m.Name); err != nil {
		return "", err
	}

	dest := filepath.Join(projectDir, ".agents", "skills", skillName(m.Name))
	if err := replaceSkillDir(dest, m); err != nil {
		return "", err
	}

	if err := copyDir(skillDir, dest); err != nil {
		return "", fmt.Errorf("install: skill: %w", err)
	}

	marker := fmt.Sprintf("plugin: %s\nversion: %s\n", m.Name, m.Version)
	if err := os.WriteFile(filepath.Join(dest, skillMarkerFile), []byte(marker), 0o644); err != nil {
		return "", fmt.Errorf("install: skill marker: %w", err)
	}
	return dest, nil
}

// replaceSkillDir empties the destination if balde owns it, errors if a
// foreign skill is already there, and does nothing when it does not exist.
func replaceSkillDir(dest string, m Manifest) error {
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(dest, skillMarkerFile))
	if err != nil {
		return fmt.Errorf("install: skill destination %s already exists and was not installed by balde", dest)
	}
	if !strings.Contains(string(data), "plugin: "+m.Name+"\n") {
		return fmt.Errorf("install: skill destination %s belongs to another plugin", dest)
	}
	return os.RemoveAll(dest)
}

// validateSkill checks the SKILL.md exists, has frontmatter with the
// required name (balde-<plugin>) and a non-empty description.
func validateSkill(path, pluginName string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("install: skill SKILL.md not found: %w", err)
	}
	content := string(data)
	if !strings.HasPrefix(content, "---") {
		return fmt.Errorf("install: skill SKILL.md must start with YAML frontmatter")
	}
	end := strings.Index(content[3:], "---")
	if end < 0 {
		return fmt.Errorf("install: skill SKILL.md frontmatter is not closed")
	}
	frontmatter := content[3 : end+3]

	wantName := "name: " + skillName(pluginName)
	if !containsLine(frontmatter, wantName) {
		return fmt.Errorf("install: skill frontmatter must declare %q", wantName)
	}
	for _, line := range strings.Split(frontmatter, "\n") {
		if strings.HasPrefix(line, "description:") && strings.TrimSpace(strings.TrimPrefix(line, "description:")) != "" {
			return nil
		}
	}
	return fmt.Errorf("install: skill frontmatter must declare a non-empty description")
}

func containsLine(block, want string) bool {
	for _, line := range strings.Split(block, "\n") {
		if strings.TrimSpace(line) == want {
			return true
		}
	}
	return false
}

func copyDir(src, dest string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// removeSkill deletes the plugin's installed skill, but only when the
// marker says balde installed it for this plugin.
func removeSkill(projectDir, pluginName string) error {
	dest := filepath.Join(projectDir, ".agents", "skills", skillName(pluginName))
	data, err := os.ReadFile(filepath.Join(dest, skillMarkerFile))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return nil // foreign skill (no marker): leave it alone
	}
	if !strings.Contains(string(data), "plugin: "+pluginName+"\n") {
		return nil
	}
	return os.RemoveAll(dest)
}
