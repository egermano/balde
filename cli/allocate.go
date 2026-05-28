package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/egermano/balde/core"
	"github.com/spf13/cobra"
)

func newAllocateCmd() *cobra.Command {
	return &cobra.Command{
		Use:  "allocate <amount> <bucket_id>",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			amount, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid amount: %s", args[0])
			}
			bucketID := args[1]

			s, err := openBudgetDB()
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer s.Close()

			budget := core.NewBudget("default", s)
			if err := budget.Allocate(bucketID, amount); err != nil {
				return fmt.Errorf("allocate: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Allocated %d cents to bucket %s\n", amount, bucketID)
			return nil
		},
	}
}

func newRainCmd() *cobra.Command {
	return &cobra.Command{
		Use:  "rain",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openBudgetDB()
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer s.Close()

			budget := core.NewBudget("default", s)
			rain, err := budget.Rain()
			if err != nil {
				return fmt.Errorf("rain: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Rain (unallocated): %d cents\n", rain)
			return nil
		},
	}
}
