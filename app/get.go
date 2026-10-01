package app

// GetBucket returns the bucket DTO for id, including archived buckets.
func (a *App) GetBucket(id string) (Bucket, error) {
	bk, err := a.store.GetBucket(id)
	if err != nil {
		return Bucket{}, err
	}
	return a.bucketDTO(bk), nil
}

// GetTransaction returns the transaction DTO for id.
func (a *App) GetTransaction(id string) (Transaction, error) {
	t, err := a.store.GetTransaction(id)
	if err != nil {
		return Transaction{}, err
	}
	return toTransactionDTO(t), nil
}
