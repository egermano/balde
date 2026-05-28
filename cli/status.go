package cli

import (
	"encoding/json"
	"fmt"

	"github.com/egermano/balde/core"
	"github.com/spf13/cobra"
)

type BudgetStatus struct {
	Accounts     []core.Account     `json:"accounts"`
	Buckets      []core.Bucket      `json:"buckets"`
	Transactions []core.Transaction `json:"transactions"`
	Rain         int64              `json:"rain"`
}

func newStatusCmd() *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:  "status",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openBudgetDB()
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer s.Close()

			budget := core.NewBudget("default", s)

			accounts, err := s.ListAccounts()
			if err != nil {
				return fmt.Errorf("list accounts: %w", err)
			}

			buckets, err := s.ListBuckets()
			if err != nil {
				return fmt.Errorf("list buckets: %w", err)
			}

			transactions, err := s.ListTransactions()
			if err != nil {
				return fmt.Errorf("list transactions: %w", err)
			}

			rain, err := budget.Rain()
			if err != nil {
				return fmt.Errorf("rain: %w", err)
			}

			status := BudgetStatus{
				Accounts:     accounts,
				Buckets:      buckets,
				Transactions: transactions,
				Rain:         rain,
			}

			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(status)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Accounts: %d\n", len(accounts))
			fmt.Fprintf(cmd.OutOrStdout(), "Buckets: %d\n", len(buckets))
			fmt.Fprintf(cmd.OutOrStdout(), "Transactions: %d\n", len(transactions))
			fmt.Fprintf(cmd.OutOrStdout(), "Rain: %d cents\n", rain)
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "output as JSON")
	return cmd
}
