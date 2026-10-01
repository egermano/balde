package app_test

import (
	"testing"
	"time"

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

func date2026January15() (t time.Time) {
	return time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
}
