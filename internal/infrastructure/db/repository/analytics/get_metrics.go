package analytics

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"log"
)

func (repo *repository) GetMetric(key []byte) ([]byte, error) {
	var metric []byte
	err := repo.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(metricsBucketName))

		metric = b.Get(key)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return metric, nil
}

func (repo *repository) GetMetrics() (map[string]entities.AnalyticsMetric, error) {
	currentMetrics := make(map[string]entities.AnalyticsMetric)
	err := repo.db.Update(func(tx *bolt.Tx) error {
		m := tx.Bucket([]byte(metricsBucketName))
		return m.ForEach(func(k, v []byte) error {
			var metric entities.AnalyticsMetric
			err := json.Unmarshal(v, &metric)
			if err != nil {
				log.Println("Error decoding api metric json from analytics db:", err)
				return nil
			}

			// add metric to currentMetrics map
			currentMetrics[metric.Date] = metric

			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return currentMetrics, nil
}
