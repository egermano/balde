package store

import "time"

func parseDate(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
