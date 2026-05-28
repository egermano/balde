package cli

import (
	"encoding/json"
	"fmt"

	"github.com/egermano/balde/store"
	"github.com/spf13/cobra"
)

func newViewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "View budget data",
	}

	cmd.AddCommand(newViewBucketsCmd())
	cmd.AddCommand(newViewTransactionsCmd())
	return cmd
}

func newViewBucketsCmd() *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:  "buckets",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.NewSQLiteStore("balde.db")
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer s.Close()

			buckets, err := s.ListBuckets()
			if err != nil {
				return fmt.Errorf("list buckets: %w", err)
			}

			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(buckets)
			}

			for _, bk := range buckets {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\ttarget=%d\tbalance=%d\n", bk.ID, bk.Name, bk.Target, bk.Balance)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "output as JSON")
	return cmd
}

func newViewTransactionsCmd() *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:  "transactions",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.NewSQLiteStore("balde.db")
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer s.Close()

			txs, err := s.ListTransactions()
			if err != nil {
				return fmt.Errorf("list transactions: %w", err)
			}

			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(txs)
			}

			for _, tx := range txs {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%d\t%s\t%s\n", tx.ID, tx.Amount, tx.Description, tx.Date.Format("2006-01-02"))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "output as JSON")
	return cmd
}
