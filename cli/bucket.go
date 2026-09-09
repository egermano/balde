package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/egermano/balde/core"
	"github.com/spf13/cobra"
)

func newBucketCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bucket",
		Short: "Manage buckets",
	}

	cmd.AddCommand(newBucketAddCmd())
	cmd.AddCommand(newBucketDeleteCmd())
	return cmd
}

func newBucketDeleteCmd() *cobra.Command {
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

			bucket, err := s.GetBucket(args[0])
			if err != nil || bucket.Archived {
				return fmt.Errorf("delete bucket: bucket not found: %s", args[0])
			}
			transactions, err := s.ListTransactions()
			if err != nil {
				return fmt.Errorf("delete bucket: %w", err)
			}
			linked := 0
			for _, transaction := range transactions {
				if transaction.BucketID == bucket.ID {
					linked++
				}
			}
			if !force && (bucket.Balance != 0 || linked != 0) {
				fmt.Fprintf(cmd.OutOrStdout(), "Archive bucket %q (id=%s, balance=%d, transactions=%d)? [y/N] ", bucket.Name, bucket.ID, bucket.Balance, linked)
				answer, _ := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
				if strings.ToLower(strings.TrimSpace(answer)) != "y" {
					fmt.Fprintln(cmd.OutOrStdout(), "Bucket not archived")
					return nil
				}
			}

			budget := core.NewBudget("default", s)
			archived, _, err := budget.DeleteBucket(bucket.ID)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Bucket %q (id=%s) archived\n", archived.Name, archived.ID)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "archive without confirmation")
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

			s, err := openBudgetDB(cmd)
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
