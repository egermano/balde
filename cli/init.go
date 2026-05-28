package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/egermano/balde/store"
	"github.com/spf13/cobra"
)

type Config struct {
	Frequency          string `json:"frequency"`
	CurrencySymbol     string `json:"currency_symbol"`
	DecimalSeparator   string `json:"decimal_separator"`
	ThousandsSeparator string `json:"thousands_separator"`
}

func DefaultConfig() Config {
	return Config{
		Frequency:          "monthly",
		CurrencySymbol:     "$",
		DecimalSeparator:   ".",
		ThousandsSeparator: ",",
	}
}

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "balde",
		Short: "Budget manager CLI using the bucket method",
	}

	root.AddCommand(newInitCmd())
	root.AddCommand(newAccountCmd())
	root.AddCommand(newBucketCmd())
	root.AddCommand(newTransactionCmd())
	root.AddCommand(newAllocateCmd())
	root.AddCommand(newRainCmd())
	root.AddCommand(newViewCmd())
	root.AddCommand(newStatusCmd())
	return root
}

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new budget",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := "balde.json"
			dbPath := "balde.db"

			if _, err := os.Stat(configPath); err == nil {
				return fmt.Errorf("budget already initialized in this directory")
			}

			s, err := store.NewSQLiteStore(dbPath)
			if err != nil {
				return fmt.Errorf("create db: %w", err)
			}
			s.Close()

			cfg := DefaultConfig()
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal config: %w", err)
			}
			if err := os.WriteFile(configPath, data, 0644); err != nil {
				return fmt.Errorf("write config: %w", err)
			}

			fmt.Println("Budget initialized.")
			return nil
		},
	}
}
