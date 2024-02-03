package analytics

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
)

var (
	requestsBucketName = "__requests"
	metricsBucketName  = "__metrics"
)

type repository struct {
	db *bolt.DB
}

func New(db *bolt.DB) (interfaces.AnalyticsRepositoryInterface, error) {
	if err := db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range []string{requestsBucketName, metricsBucketName} {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to create analytics storage buckets: %v", err)
	}

	return &repository{db: db}, nil
}
