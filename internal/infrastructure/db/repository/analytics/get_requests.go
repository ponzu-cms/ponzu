package analytics

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"log"
	"time"
)

func (repo *repository) GetRequestMetadata(
	t time.Time,
	currentMetrics map[string]entities.AnalyticsMetric,
) ([]entities.AnalyticsHTTPRequestMetadata, error) {
	metrics := currentMetrics
	var err error
	if metrics == nil {
		metrics, err = repo.GetMetrics()
		if err != nil {
			return nil, err
		}
	}

	requests := make([]entities.AnalyticsHTTPRequestMetadata, 0)
	err = repo.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(requestsBucketName))

		return b.ForEach(func(k, v []byte) error {
			var r entities.AnalyticsHTTPRequestMetadata
			err := json.Unmarshal(v, &r)
			if err != nil {
				log.Println("Error decoding api request json from analytics db:", err)
				return nil
			}

			// append request to requests for analysis if its timestamp is t
			// or if its day is not already in cache, otherwise delete it
			d := time.Unix(r.Timestamp/1000, 0)
			_, inCache := metrics[d.Format("01/02")]
			if !d.Before(t) || !inCache {
				requests = append(requests, r)
			} else {
				err := b.Delete(k)
				if err != nil {
					return err
				}
			}

			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return requests, nil
}
