package app

// ListBuckets returns all non-archived buckets as DTOs. Empty result is a
// non-nil empty slice.
func (a *App) ListBuckets() ([]Bucket, error) {
	buckets, err := a.store.ListBuckets()
	if err != nil {
		return nil, err
	}
	out := make([]Bucket, 0, len(buckets))
	for _, bk := range buckets {
		out = append(out, toBucketDTO(bk))
	}
	return out, nil
}
