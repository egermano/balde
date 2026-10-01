package cli

import (
	"fmt"

	"github.com/egermano/balde/app"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:  "status",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openBudgetDB(cmd)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer s.Close()

			a := app.New("default", s)
			status, err := a.Status()
			if err != nil {
				return fmt.Errorf("status: %w", err)
			}

			if asJSON {
				return writeJSON(cmd, status)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Accounts: %d\n", len(status.Accounts))
			fmt.Fprintf(cmd.OutOrStdout(), "Buckets: %d\n", len(status.Buckets))
			fmt.Fprintf(cmd.OutOrStdout(), "Transactions: %d\n", len(status.Transactions))
			fmt.Fprintf(cmd.OutOrStdout(), "Rain: %d cents\n", status.Rain)
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "output as JSON")
	return cmd
}
