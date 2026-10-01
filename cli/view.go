package cli

import (
	"fmt"

	"github.com/egermano/balde/app"
	"github.com/egermano/balde/core"
	"github.com/egermano/balde/store"
	"github.com/spf13/cobra"
)

func newViewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "View budget data",
	}

	cmd.AddCommand(newViewBucketsCmd())
	cmd.AddCommand(newViewAccountsCmd())
	cmd.AddCommand(newViewTransactionsCmd())
	return cmd
}

func newViewAccountsCmd() *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:  "accounts",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openBudgetDB(cmd)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer s.Close()

			a := app.New("default", s)
			accounts, err := a.ListAccounts()
			if err != nil {
				return fmt.Errorf("list accounts: %w", err)
			}
			if asJSON {
				return writeJSON(cmd, accounts)
			}

			dbPath, err := budgetDBPath(cmd)
			if err != nil {
				return err
			}
			config, err := store.ReadConfig(dbPath)
			if err != nil {
				return fmt.Errorf("read config: %w", err)
			}

			for _, account := range accounts {
				balance := core.FormatAmount(account.Balance, config.DecimalSeparator, config.ThousandsSeparator, config.CurrencySymbol)
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\n", account.ID, account.Name, account.Type, balance)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "output as JSON")
	return cmd
}

func newViewBucketsCmd() *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:  "buckets",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openBudgetDB(cmd)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer s.Close()

			a := app.New("default", s)
			buckets, err := a.ListBuckets()
			if err != nil {
				return fmt.Errorf("list buckets: %w", err)
			}

			if asJSON {
				return writeJSON(cmd, buckets)
			}

			for _, bk := range buckets {
				fillDisplay := "Not set"
				if bk.Target > 0 {
					fillDisplay = fmt.Sprintf("%.1f%%", bk.FillPercent)
				}

				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\ttarget=%d\tbalance=%d\tfill=%s\n",
					bk.ID, bk.Name, bk.Target, bk.Balance, fillDisplay)
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
			s, err := openBudgetDB(cmd)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer s.Close()

			a := app.New("default", s)
			txs, err := a.ListTransactions()
			if err != nil {
				return fmt.Errorf("list transactions: %w", err)
			}

			if asJSON {
				return writeJSON(cmd, txs)
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
