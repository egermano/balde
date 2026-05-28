package cli

import (
	"fmt"
	"os"

	"github.com/egermano/balde/core"
	"github.com/egermano/balde/store"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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
	root.AddCommand(newUnlockCmd())
	root.AddCommand(newLockCmd())
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
	var password string
	var dir string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new budget",
		RunE: func(cmd *cobra.Command, args []string) error {
			origDir, _ := os.Getwd()
			defer os.Chdir(origDir)

			if dir != "" {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return fmt.Errorf("create directory: %w", err)
				}
				if err := os.Chdir(dir); err != nil {
					return fmt.Errorf("change directory: %w", err)
				}
			}

			configPath := "balde.json"
			dbPath := "balde.db"

			if _, err := os.Stat(configPath); err == nil {
				return fmt.Errorf("budget already initialized in this directory")
			}

			var s store.Store
			var err error

			encrypted := false
			if password != "" {
				encrypted = true
			} else {
				fmt.Print("Enable encryption? (y/N): ")
				var response string
				fmt.Scanln(&response)
				if response == "y" || response == "Y" {
					encrypted = true

					fmt.Print("Enter password: ")
					bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
					if err != nil {
						return fmt.Errorf("read password: %w", err)
					}
					password = string(bytePassword)
					fmt.Println()

					fmt.Print("Confirm password: ")
					bytePassword2, err := term.ReadPassword(int(os.Stdin.Fd()))
					if err != nil {
						return fmt.Errorf("read password: %w", err)
					}
					password2 := string(bytePassword2)
					fmt.Println()

					if password != password2 {
						return fmt.Errorf("passwords do not match")
					}
				}
			}

			cfg := DefaultConfig()

			if encrypted {
				cfgEnc := store.Config{
					Frequency:          cfg.Frequency,
					CurrencySymbol:    cfg.CurrencySymbol,
					DecimalSeparator:  cfg.DecimalSeparator,
					ThousandsSeparator: cfg.ThousandsSeparator,
					Encrypted:         true,
				}
				if err := store.WriteConfig(dbPath, cfgEnc); err != nil {
					return fmt.Errorf("write config: %w", err)
				}

				s, err = store.OpenStore(dbPath, password, os.Getenv("BALDE_PASSWORD"), cfgEnc)
			} else {
				cfgPlain := store.Config{
					Frequency:          cfg.Frequency,
					CurrencySymbol:    cfg.CurrencySymbol,
					DecimalSeparator:  cfg.DecimalSeparator,
					ThousandsSeparator: cfg.ThousandsSeparator,
					Encrypted:         false,
				}
				if err := store.WriteConfig(dbPath, cfgPlain); err != nil {
					return fmt.Errorf("write config: %w", err)
				}

				s, err = store.NewSQLiteStore(dbPath)
			}

			if err != nil {
				return fmt.Errorf("create db: %w", err)
			}
			defer s.Close()

			budget := core.NewBudget("default", s)
			defaultBuckets := []struct {
				name   string
				target int64
			}{
				{"financial freedom", 0},
				{"fixed costs", 0},
				{"pleasures", 0},
				{"comfort", 0},
				{"knowledge", 0},
				{"goals", 0},
			}
			for _, db := range defaultBuckets {
				if _, err := budget.AddBucket(db.name, db.target); err != nil {
					return fmt.Errorf("create default bucket %s: %w", db.name, err)
				}
			}

			fmt.Println("Budget initialized.")
			return nil
		},
	}

	cmd.Flags().StringVarP(&password, "password", "p", "", "Password for encryption")
	cmd.Flags().StringVarP(&dir, "dir", "d", "", "Directory to initialize budget in")

	return cmd
}

func newUnlockCmd() *cobra.Command {
	var password string
	var dbPath string

	cmd := &cobra.Command{
		Use:   "unlock",
		Short: "Unlock an encrypted budget database",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				dbPath = "balde.db"
			}

			config, err := store.ReadConfig(dbPath)
			if err != nil {
				return fmt.Errorf("read config: %w", err)
			}

			if !config.Encrypted {
				return fmt.Errorf("database is not encrypted")
			}

			envPassword := os.Getenv("BALDE_PASSWORD")
			if password == "" && envPassword == "" {
				return fmt.Errorf("password required (use --password flag or BALDE_PASSWORD env var)")
			}

			if password == "" {
				password = envPassword
			}

			s, err := store.OpenStore(dbPath, password, "", config)
			if err != nil {
				return fmt.Errorf("unlock failed: %w", err)
			}
			defer s.Close()

			sessionPath := store.GetSessionPath(dbPath)
			fmt.Printf("Successfully unlocked database. Session valid for 30 minutes.\n")
			fmt.Printf("Session file: %s\n", sessionPath)

			return nil
		},
	}

	cmd.Flags().StringVarP(&password, "password", "p", "", "Password for encryption")
	cmd.Flags().StringVarP(&dbPath, "db", "", "balde.db", "Path to budget database")

	return cmd
}

func newLockCmd() *cobra.Command {
	var dbPath string

	cmd := &cobra.Command{
		Use:   "lock",
		Short: "Lock an encrypted budget database by invalidating session",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				dbPath = "balde.db"
			}

			config, err := store.ReadConfig(dbPath)
			if err != nil {
				return fmt.Errorf("read config: %w", err)
			}

			if !config.Encrypted {
				return fmt.Errorf("database is not encrypted")
			}

			sessionPath := store.GetSessionPath(dbPath)
			if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
				return fmt.Errorf("no active session")
			}

			if err := os.Remove(sessionPath); err != nil {
				return fmt.Errorf("remove session: %w", err)
			}

			fmt.Println("Database locked. Session invalidated.")
			return nil
		},
	}

	cmd.Flags().StringVarP(&dbPath, "db", "", "balde.db", "Path to budget database")

	return cmd
}
