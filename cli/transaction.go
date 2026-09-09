package cli

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/egermano/balde/core"
	"github.com/spf13/cobra"
)

func newTransactionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transaction",
		Short: "Manage transactions",
	}

	cmd.AddCommand(newTransactionAddCmd())
	cmd.AddCommand(newTransactionDeleteCmd())
	return cmd
}

func newTransactionAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "add <amount> <description> <account_id> <bucket_id>",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 4 {
				return cobra.ExactArgs(4)(cmd, args)
			}
			if err := cmd.Flags().Parse(args[4:]); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			amount, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid amount: %s", args[0])
			}
			description := args[1]
			accountID := args[2]
			bucketID := args[3]

			s, err := openBudgetDB(cmd)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer s.Close()

			budget := core.NewBudget("default", s)
			tx, err := budget.AddTransaction(amount, description, time.Now(), accountID, bucketID)
			if err != nil {
				return fmt.Errorf("add transaction: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Transaction created: amount=%d desc=%s id=%s\n", tx.Amount, tx.Description, tx.ID)
			return nil
		},
	}
	cmd.DisableFlagParsing = true
	return cmd
}

func newTransactionDeleteCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:  "delete <id>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openBudgetDB(cmd)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer s.Close()

			tx, err := s.GetTransaction(args[0])
			if err != nil {
				return fmt.Errorf("delete transaction: %w", err)
			}
			if !force {
				fmt.Fprintf(cmd.OutOrStdout(), "Delete transaction %s (%d)? [y/N] ", tx.Description, tx.Amount)
				answer, _ := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
				if strings.ToLower(strings.TrimSpace(answer)) != "y" {
					fmt.Fprintln(cmd.OutOrStdout(), "Transaction not deleted")
					return nil
				}
			}

			budget := core.NewBudget("default", s)
			deleted, err := budget.DeleteTransaction(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Transaction deleted: %s (%d)\n", deleted.Description, deleted.Amount)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "delete without confirmation")
	return cmd
}
