package cli

import (
	"fmt"
	"os"

	"github.com/egermano/balde/plugin"
	"github.com/spf13/cobra"
)

func newPluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage balde plugins (experimental)",
	}

	cmd.AddCommand(newPluginInstallCmd())
	cmd.AddCommand(newPluginListCmd())
	cmd.AddCommand(newPluginRemoveCmd())
	return cmd
}

// experimentalEnabled reports whether the user opted into experimental
// features via the --experimental flag or BALDE_EXPERIMENTAL=1.
func experimentalEnabled(cmd *cobra.Command) bool {
	if v := os.Getenv("BALDE_EXPERIMENTAL"); v == "1" {
		return true
	}
	if v, err := cmd.Root().PersistentFlags().GetBool("experimental"); err == nil && v {
		return true
	}
	v, err := cmd.Flags().GetBool("experimental")
	return err == nil && v
}

// requireExperimental gates plugin commands behind the experimental opt-in.
func requireExperimental(cmd *cobra.Command) error {
	if experimentalEnabled(cmd) {
		fmt.Fprintln(cmd.ErrOrStderr(), "notice: plugin support is experimental and may change or break without notice")
		return nil
	}
	return fmt.Errorf("plugin commands are experimental: opt in with --experimental or BALDE_EXPERIMENTAL=1")
}

// projectDir resolves the current budget directory (--dir or CWD).
func projectDir(cmd *cobra.Command) (string, error) {
	dir, err := cmd.Flags().GetString("dir")
	if err != nil {
		return "", err
	}
	return dirOrDefault(dir), nil
}

func newPluginInstallCmd() *cobra.Command {
	var path string
	cmd := &cobra.Command{
		Use:   "install <source>",
		Args:  cobra.ExactArgs(1),
		Short: "Install a plugin from a git repository (path or URL)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireExperimental(cmd); err != nil {
				return err
			}

			dir, err := projectDir(cmd)
			if err != nil {
				return err
			}

			entry, err := plugin.Install(dir, args[0], path)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Plugin installed: %s %s (commit %s)\n", entry.Name, entry.Version, shortSHA(entry.SHA))
			if entry.HasSkill() {
				fmt.Fprintf(cmd.OutOrStdout(), "Skill installed: .agents/skills/balde-%s\n", entry.Name)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "subdirectory inside the repository where the plugin lives")
	return cmd
}

func newPluginListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Short: "List installed plugins and their permissions",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireExperimental(cmd); err != nil {
				return err
			}

			dir, err := projectDir(cmd)
			if err != nil {
				return err
			}

			lock, err := plugin.ReadLockfile(dir)
			if err != nil {
				return fmt.Errorf("plugin list: %w", err)
			}
			if len(lock.Plugins) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No plugins installed")
				return nil
			}
			for _, p := range lock.Plugins {
				skill := "-"
				if p.HasSkill() {
					skill = "balde-" + p.Name
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\tskill=%s\n", p.Name, p.Version, p.Source, skill)
			}
			return nil
		},
	}
}

func newPluginRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Args:  cobra.ExactArgs(1),
		Short: "Uninstall a plugin",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireExperimental(cmd); err != nil {
				return err
			}

			dir, err := projectDir(cmd)
			if err != nil {
				return err
			}

			if err := plugin.Remove(dir, args[0]); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Plugin removed: %s\n", args[0])
			return nil
		},
	}
}

func dirOrDefault(dir string) string {
	if dir == "" {
		return "."
	}
	return dir
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
