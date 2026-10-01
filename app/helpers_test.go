package app_test

import (
	"testing"

	"github.com/egermano/balde/core"
)

func mustCreateAccount(t *testing.T, s core.Store, a core.Account) {
	t.Helper()
	if err := s.CreateAccount(a); err != nil {
		t.Fatalf("seed account: %v", err)
	}
}

func mustCreateBucket(t *testing.T, s core.Store, b core.Bucket) {
	t.Helper()
	if err := s.CreateBucket(b); err != nil {
		t.Fatalf("seed bucket: %v", err)
	}
}

func mustCreateTransaction(t *testing.T, s core.Store, tr core.Transaction) {
	t.Helper()
	if err := s.CreateTransaction(tr); err != nil {
		t.Fatalf("seed transaction: %v", err)
	}
}
