package store

import (
	"fmt"

	"github.com/egermano/balde/core"
)

// rawInserter is implemented by stores that can accept verbatim rows
// (explicit IDs, no side effects). Budget migrations need it because the
// regular Create* methods assign fresh IDs and apply transaction deltas to
// balances — re-applying them corrupts the copied budget (#51).
type rawInserter interface {
	insertAccountRaw(core.Account) error
	insertBucketRaw(core.Bucket) error
	insertTransactionRaw(core.Transaction) error
}

// CopyBudget copies every row from src to dst verbatim: same IDs (so
// transaction references stay valid), same balances (no re-applied
// deltas), archived buckets included. It does not persist dst — call Close
// to save, or Discard to throw the copy away.
func CopyBudget(src, dst core.Store) error {
	ri, ok := dst.(rawInserter)
	if !ok {
		return fmt.Errorf("copy: destination store %T does not support raw migration", dst)
	}

	accounts, err := src.ListAccounts()
	if err != nil {
		return fmt.Errorf("copy: list accounts: %w", err)
	}
	for _, a := range accounts {
		if err := ri.insertAccountRaw(a); err != nil {
			return fmt.Errorf("copy: account %s: %w", a.ID, err)
		}
	}

	buckets, err := src.ListAllBuckets()
	if err != nil {
		return fmt.Errorf("copy: list buckets: %w", err)
	}
	for _, b := range buckets {
		if err := ri.insertBucketRaw(b); err != nil {
			return fmt.Errorf("copy: bucket %s: %w", b.ID, err)
		}
	}

	transactions, err := src.ListTransactions()
	if err != nil {
		return fmt.Errorf("copy: list transactions: %w", err)
	}
	for _, t := range transactions {
		if err := ri.insertTransactionRaw(t); err != nil {
			return fmt.Errorf("copy: transaction %s: %w", t.ID, err)
		}
	}
	return nil
}
