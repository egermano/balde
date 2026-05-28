package core

type Store interface {
	CreateAccount(a Account) error
	GetAccount(id string) (Account, error)
	ListAccounts() ([]Account, error)
	UpdateAccount(a Account) error

	CreateBucket(b Bucket) error
	GetBucket(id string) (Bucket, error)
	ListBuckets() ([]Bucket, error)
	UpdateBucket(b Bucket) error
	DeleteBucket(id string) error

	CreateTransaction(t Transaction) error
	GetTransaction(id string) (Transaction, error)
	ListTransactions() ([]Transaction, error)
	UpdateTransaction(t Transaction) error
}
