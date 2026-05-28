package cli

import (
	"fmt"
	"os"

	"github.com/egermano/balde/store"
	"golang.org/x/term"
)

func openBudgetDB() (store.Store, error) {
	dbPath := "balde.db"

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
