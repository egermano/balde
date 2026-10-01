package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/egermano/balde/app"
	"github.com/egermano/balde/core"
	"github.com/spf13/cobra"
)

func newAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Manage accounts",
	}

	cmd.AddCommand(newAccountAddCmd())
	return cmd
}

func newAccountAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:  "add <name> <type> <balance>",
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			accountType := core.AccountType(args[1])
			balance, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid balance: %s", args[2])
			}

			s, err := openBudgetDB(cmd)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer s.Close()

			a := app.New("default", s)
			acc, err := a.AddAccount(name, accountType, balance)
			if err != nil {
				return fmt.Errorf("add account: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Account created: %s (%s) balance=%d id=%s\n", acc.Name, acc.Type, acc.Balance, acc.ID)
			return nil
		},
	}
}
