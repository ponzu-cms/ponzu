package analytics

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"time"
)

// Prune takes a duration to evaluate apiRequest dates against. If any of
// the apiRequest timestamps are before the threshold, they are removed.
// TODO: add feature to alternatively backup old analytics to cloud
func (repo *repository) Prune(threshold time.Duration) error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	maxTimeToUpdate := today.Add(threshold)

	// iterate through all request data
	err := repo.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(requestsBucketName))

		err := b.ForEach(func(k, v []byte) error {
			var r entities.AnalyticsHTTPRequestMetadata
			err := json.Unmarshal(v, &r)
			if err != nil {
				return err
			}

			// delete if timestamp is below or equal to maxTimeToUpdate
			ts := time.Unix(r.Timestamp/1000, 0)
			if ts.Equal(maxTimeToUpdate) || ts.Before(maxTimeToUpdate) {
				err := b.Delete(k)
				if err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
