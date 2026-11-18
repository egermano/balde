package core

type Store interface {
	// Create* persists the entity, assigns its ID, and returns the stored
	// entity. Implementations must never make callers guess which row was
	// created (list ordering is not insertion order).
	CreateAccount(a Account) (Account, error)
	GetAccount(id string) (Account, error)
	ListAccounts() ([]Account, error)
	UpdateAccount(a Account) error

	CreateBucket(b Bucket) (Bucket, error)
	GetBucket(id string) (Bucket, error)
	ListBuckets() ([]Bucket, error)
	// ListAllBuckets includes archived buckets — used by migrations that
	// must copy every row; normal flows use ListBuckets.
	ListAllBuckets() ([]Bucket, error)
	UpdateBucket(b Bucket) error
	DeleteBucket(id string) error

	CreateTransaction(t Transaction) (Transaction, error)
	GetTransaction(id string) (Transaction, error)
	ListTransactions() ([]Transaction, error)
	DeleteTransaction(id string) error
}
