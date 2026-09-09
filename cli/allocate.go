package cli

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/egermano/balde/core"
	"github.com/spf13/cobra"
)

func newAllocateCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:  "allocate <amount> <bucket_id>",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			amount, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid amount: %s", args[0])
			}
			bucketID := args[1]

			s, err := openBudgetDB(cmd)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer s.Close()

			budget := core.NewBudget("default", s)
			if _, err := s.GetBucket(bucketID); err != nil {
				return fmt.Errorf("allocate: %w", err)
			}
			if amount > 0 && !force {
				rain, err := budget.Rain()
				if err != nil {
					return fmt.Errorf("rain: %w", err)
				}
				if amount > rain {
					fmt.Fprintf(cmd.OutOrStdout(), "Warning: You have %d cents available, but trying to allocate %d cents.\nThis will result in negative rain. Continue? (y/N) ", rain, amount)
					answer, _ := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
					answer = strings.TrimSpace(answer)
					if !strings.EqualFold(answer, "y") && !strings.EqualFold(answer, "yes") {
						return nil
					}
				}
			}
			if err := budget.Allocate(bucketID, amount); err != nil {
				return fmt.Errorf("allocate: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Allocated %d cents to bucket %s\n", amount, bucketID)
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Allocate without confirmation")
	return cmd
}

func newRainCmd() *cobra.Command {
	return &cobra.Command{
		Use:  "rain",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openBudgetDB(cmd)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer s.Close()

			budget := core.NewBudget("default", s)
			rain, err := budget.Rain()
			if err != nil {
				return fmt.Errorf("rain: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Rain (unallocated): %d cents\n", rain)
			return nil
		},
	}
}
