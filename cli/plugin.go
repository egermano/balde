package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newPluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage balde plugins (experimental)",
	}
	cmd.PersistentFlags().Bool("experimental", false, "opt in to experimental features")

	cmd.AddCommand(newPluginListCmd())
	return cmd
}

// experimentalEnabled reports whether the user opted into experimental
// features via the --experimental flag or BALDE_EXPERIMENTAL=1.
func experimentalEnabled(cmd *cobra.Command) bool {
	if os.Getenv("BALDE_EXPERIMENTAL") == "1" {
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

func newPluginListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Short: "List installed plugins and their permissions",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireExperimental(cmd); err != nil {
				return err
			}

			dir, err := cmd.Flags().GetString("dir")
			if err != nil {
				return err
			}
			lockPath := filepath.Join(dirOrDefault(dir), ".balde", "plugins", "lock.json")

			data, err := os.ReadFile(lockPath)
			if os.IsNotExist(err) {
				fmt.Fprintln(cmd.OutOrStdout(), "No plugins installed")
				return nil
			}
			if err != nil {
				return fmt.Errorf("read lockfile: %w", err)
			}

			var lock struct {
				Plugins []struct {
					Name    string `json:"name"`
					Version string `json:"version"`
					Source  string `json:"source"`
				} `json:"plugins"`
			}
			if err := json.Unmarshal(data, &lock); err != nil {
				return fmt.Errorf("parse lockfile: %w", err)
			}
			for _, p := range lock.Plugins {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", p.Name, p.Version, p.Source)
			}
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
