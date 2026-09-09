package cli

import (
	"fmt"
	"os"

	"github.com/egermano/balde/store"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func openBudgetDB(cmd *cobra.Command) (store.Store, error) {
	dbPath, err := budgetDBPath(cmd)
	if err != nil {
		return nil, err
	}

	config, err := store.ReadConfig(dbPath)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	envPassword := os.Getenv("BALDE_PASSWORD")

	s, err := store.OpenStore(dbPath, "", envPassword, config)
	if err == nil {
		return s, nil
	}

	if !config.Encrypted {
		return nil, err
	}

	if envPassword != "" {
		return nil, fmt.Errorf("unlock failed: %w", err)
	}

	fmt.Print("Enter password: ")
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("read password: %w", err)
	}
	password := string(bytePassword)
	fmt.Println()

	return store.OpenStore(dbPath, password, "", config)
}

func budgetDBPath(cmd *cobra.Command) (string, error) {
	dir, err := cmd.Flags().GetString("dir")
	if err != nil {
		return "", err
	}
	dbPath := "balde.db"
	if dir != "" {
		dbPath = dir + string(os.PathSeparator) + dbPath
	}
	return dbPath, nil
}
