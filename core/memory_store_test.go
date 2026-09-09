package core_test

import (
	"fmt"
	"sync"

	"github.com/egermano/balde/core"
)

type MemoryStore struct {
	mu           sync.RWMutex
	accounts     map[string]core.Account
	buckets      map[string]core.Bucket
	transactions map[string]core.Transaction
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		accounts:     make(map[string]core.Account),
		buckets:      make(map[string]core.Bucket),
		transactions: make(map[string]core.Transaction),
	}
}

func (m *MemoryStore) CreateAccount(a core.Account) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if a.ID == "" {
		a.ID = fmt.Sprintf("acc-%d", len(m.accounts)+1)
	}
	m.accounts[a.ID] = a
	return nil
}

func (m *MemoryStore) GetAccount(id string) (core.Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.accounts[id]
	if !ok {
		return core.Account{}, fmt.Errorf("account not found: %s", id)
	}
	return a, nil
}

func (m *MemoryStore) ListAccounts() ([]core.Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]core.Account, 0, len(m.accounts))
	for _, a := range m.accounts {
		result = append(result, a)
	}
	return result, nil
}

func (m *MemoryStore) UpdateAccount(a core.Account) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.accounts[a.ID]; !ok {
		return fmt.Errorf("account not found: %s", a.ID)
	}
	m.accounts[a.ID] = a
	return nil
}

func (m *MemoryStore) CreateBucket(b core.Bucket) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if b.ID == "" {
		b.ID = fmt.Sprintf("bkt-%d", len(m.buckets)+1)
	}
	m.buckets[b.ID] = b
	return nil
}

func (m *MemoryStore) GetBucket(id string) (core.Bucket, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	b, ok := m.buckets[id]
	if !ok {
		return core.Bucket{}, fmt.Errorf("bucket not found: %s", id)
	}
	return b, nil
}

func (m *MemoryStore) ListBuckets() ([]core.Bucket, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]core.Bucket, 0, len(m.buckets))
	for _, b := range m.buckets {
		result = append(result, b)
	}
	return result, nil
}

func (m *MemoryStore) UpdateBucket(b core.Bucket) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.buckets[b.ID]; !ok {
		return fmt.Errorf("bucket not found: %s", b.ID)
	}
	m.buckets[b.ID] = b
	return nil
}

func (m *MemoryStore) DeleteBucket(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.buckets, id)
	return nil
}

func (m *MemoryStore) CreateTransaction(t core.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t.ID == "" {
		t.ID = fmt.Sprintf("txn-%d", len(m.transactions)+1)
	}
	m.transactions[t.ID] = t
	return nil
}

func (m *MemoryStore) GetTransaction(id string) (core.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.transactions[id]
	if !ok {
		return core.Transaction{}, fmt.Errorf("transaction not found: %s", id)
	}
	return t, nil
}

func (m *MemoryStore) ListTransactions() ([]core.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]core.Transaction, 0, len(m.transactions))
	for _, t := range m.transactions {
		result = append(result, t)
	}
	return result, nil
}

func (m *MemoryStore) UpdateTransaction(t core.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.transactions[t.ID]; !ok {
		return fmt.Errorf("transaction not found: %s", t.ID)
	}
	m.transactions[t.ID] = t
	return nil
}
