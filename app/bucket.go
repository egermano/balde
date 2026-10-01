package app

// AddBucket creates a bucket and returns its DTO.
func (a *App) AddBucket(name string, target int64) (Bucket, error) {
	bk, err := a.budget.AddBucket(name, target)
	if err != nil {
		return Bucket{}, err
	}
	return a.bucketDTO(bk), nil
}

// DeleteBucket archives the bucket and returns its DTO plus the number of
// transactions linked to it.
func (a *App) DeleteBucket(id string) (Bucket, int, error) {
	bucket, linked, err := a.budget.DeleteBucket(id)
	if err != nil {
		return Bucket{}, 0, err
	}
	return a.bucketDTO(bucket), linked, nil
}

// ListBuckets returns all non-archived buckets as DTOs. Empty result is a
// non-nil empty slice.
func (a *App) ListBuckets() ([]Bucket, error) {
	buckets, err := a.store.ListBuckets()
	if err != nil {
		return nil, err
	}
	out := make([]Bucket, 0, len(buckets))
	for _, bk := range buckets {
		out = append(out, a.bucketDTO(bk))
	}
	return out, nil
}
