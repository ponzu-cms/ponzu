package analytics

import "github.com/boltdb/bolt"

func (repo *repository) SetMetric(key, value []byte) error {
	return repo.db.Update(func(tx *bolt.Tx) error {
		m := tx.Bucket([]byte(metricsBucketName))

		return m.Put(key, value)
	})
}
