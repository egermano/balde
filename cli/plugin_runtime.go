package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/egermano/balde/app"
	"github.com/egermano/balde/plugin"
	"github.com/egermano/balde/store"
	"github.com/spf13/cobra"
)

// registerInstalledPluginCommands makes project-local command capabilities
// available as normal root commands. A malformed or absent lockfile leaves
// the base CLI usable; `balde plugin list` reports lockfile errors directly.
func registerInstalledPluginCommands(root *cobra.Command) {
	lock, err := plugin.ReadLockfile(pluginProjectDirFromArgs(os.Args[1:]))
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
			pluginArgs, err := consumePluginHostFlags(cmd, args)
			if err != nil {
				return err
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
			forwardedArgs := make([]string, 0, len(pluginArgs))
			for _, arg := range pluginArgs {
				if arg == "--json" {
					asJSON = true
					continue
				}
				forwardedArgs = append(forwardedArgs, arg)
			}
			result, err := runner.Run(capability.Name, forwardedArgs)
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

// pluginProjectDirFromArgs mirrors the persistent --dir flag early enough to
// discover command capabilities, before Cobra resolves the requested command.
func pluginProjectDirFromArgs(args []string) string {
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--dir" || args[i] == "-d":
			if i+1 < len(args) {
				return dirOrDefault(args[i+1])
			}
		case strings.HasPrefix(args[i], "--dir="):
			return dirOrDefault(strings.TrimPrefix(args[i], "--dir="))
		}
	}
	return "."
}

// consumePluginHostFlags handles inherited host flags that Cobra intentionally
// leaves opaque so plugin-owned flags can pass through unchanged.
func consumePluginHostFlags(cmd *cobra.Command, args []string) ([]string, error) {
	forwarded := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--experimental":
			if err := cmd.Root().PersistentFlags().Set("experimental", "true"); err != nil {
				return nil, err
			}
			continue
		case "--dir", "-d":
			if i+1 == len(args) {
				return nil, fmt.Errorf("flag needs an argument: %s", args[i])
			}
			if err := cmd.Root().PersistentFlags().Set("dir", args[i+1]); err != nil {
				return nil, err
			}
			i++
			continue
		}
		if strings.HasPrefix(args[i], "--dir=") {
			if err := cmd.Root().PersistentFlags().Set("dir", strings.TrimPrefix(args[i], "--dir=")); err != nil {
				return nil, err
			}
			continue
		}
		forwarded = append(forwarded, args[i])
	}
	return forwarded, nil
}
