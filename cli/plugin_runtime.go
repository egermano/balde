package cli

import (
	"fmt"

	"github.com/egermano/balde/app"
	"github.com/egermano/balde/plugin"
	"github.com/egermano/balde/store"
	"github.com/spf13/cobra"
)

// registerInstalledPluginCommands makes project-local command capabilities
// available as normal root commands. A malformed or absent lockfile leaves
// the base CLI usable; `balde plugin list` reports lockfile errors directly.
func registerInstalledPluginCommands(root *cobra.Command) {
	lock, err := plugin.ReadLockfile(".")
	if err != nil {
		return
	}

	reserved := make(map[string]bool, len(root.Commands()))
	for _, cmd := range root.Commands() {
		reserved[cmd.Name()] = true
	}
	for _, entry := range lock.Plugins {
		if entry.Run == "" {
			continue // lockfiles written before the runtime are not runnable
		}
		for _, capability := range entry.Capabilities {
			if capability.Type != plugin.CapCommand || capability.Name == "" || reserved[capability.Name] {
				continue
			}
			root.AddCommand(newInstalledPluginCmd(entry, capability))
			reserved[capability.Name] = true
		}
	}
}

func newInstalledPluginCmd(entry plugin.LockEntry, capability plugin.Capability) *cobra.Command {
	cmd := &cobra.Command{
		Use:   capability.Name + " [args...]",
		Short: capability.Description,
		// Plugin-owned flags (--date, --budget, …) are intentionally opaque to
		// Cobra. The host only recognizes --json and passes all other args on.
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// DisableFlagParsing leaves even inherited flags in args. Consume the
			// experimental opt-in here before applying the runtime gate.
			for _, arg := range args {
				if arg == "--experimental" {
					if err := cmd.Root().PersistentFlags().Set("experimental", "true"); err != nil {
						return err
					}
				}
			}
			if err := requireExperimental(cmd); err != nil {
				return err
			}

			s, err := openBudgetDB(cmd)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer s.Close()

			dbPath, err := budgetDBPath(cmd)
			if err != nil {
				return err
			}
			config, err := store.ReadConfig(dbPath)
			if err != nil {
				return fmt.Errorf("read config: %w", err)
			}
			dir, err := projectDir(cmd)
			if err != nil {
				return err
			}

			runner := &plugin.Runner{
				Entry:  entry,
				SrcDir: plugin.SourceDir(dir, entry.Name),
				App:    app.New("default", s),
				Meta: plugin.BudgetMeta{
					CurrencySymbol:     config.CurrencySymbol,
					DecimalSeparator:   config.DecimalSeparator,
					ThousandsSeparator: config.ThousandsSeparator,
					Frequency:          config.Frequency,
				},
				Stderr: cmd.ErrOrStderr(),
			}
			asJSON := false
			pluginArgs := make([]string, 0, len(args))
			for _, arg := range args {
				if arg == "--json" {
					asJSON = true
					continue
				}
				if arg == "--experimental" {
					continue
				}
				pluginArgs = append(pluginArgs, arg)
			}
			result, err := runner.Run(capability.Name, pluginArgs)
			if err != nil {
				return err
			}

			if asJSON {
				if len(result.JSON) > 0 {
					if _, err := cmd.OutOrStdout().Write(append(result.JSON, '\n')); err != nil {
						return err
					}
					return nil
				}
				return writeJSON(cmd, map[string]string{"text": result.Text})
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), result.Text)
			return err
		},
	}
	return cmd
}
