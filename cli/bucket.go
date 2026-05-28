package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/egermano/balde/core"
	"github.com/spf13/cobra"
)

func newBucketCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bucket",
		Short: "Manage buckets",
	}

	cmd.AddCommand(newBucketAddCmd())
	return cmd
}

func newBucketAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:  "add <name> <target>",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			target, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid target: %s", args[1])
			}

			s, err := openBudgetDB()
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer s.Close()

			budget := core.NewBudget("default", s)
			bk, err := budget.AddBucket(name, target)
			if err != nil {
				return fmt.Errorf("add bucket: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Bucket created: %s target=%d id=%s\n", bk.Name, bk.Target, bk.ID)
			return nil
		},
	}
}
