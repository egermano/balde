package cli

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/egermano/balde/core"
	"github.com/egermano/balde/store"
	"github.com/spf13/cobra"
)

func newTransactionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transaction",
		Short: "Manage transactions",
	}

	cmd.AddCommand(newTransactionAddCmd())
	return cmd
}

func newTransactionAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "add <amount> <description> <account_id> <bucket_id>",
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			amount, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid amount: %s", args[0])
			}
			description := args[1]
			accountID := args[2]
			bucketID := args[3]

			s, err := store.NewSQLiteStore("balde.db")
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer s.Close()

			budget := core.NewBudget("default", s)
			tx, err := budget.AddTransaction(amount, description, time.Now(), accountID, bucketID)
			if err != nil {
				return fmt.Errorf("add transaction: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Transaction created: amount=%d desc=%s id=%s\n", tx.Amount, tx.Description, tx.ID)
			return nil
		},
	}
	cmd.DisableFlagParsing = true
	return cmd
}
