package analytics

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"strconv"
)

func (repo *repository) Insert(reqs []entities.AnalyticsHTTPRequestMetadata) error {
	err := repo.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("__requests"))
		if err != nil {
			return err
		}

		for _, apiReq := range reqs {
			// get the next available ID and convert to string
			// also set effectedID to int of ID
			id, err := b.NextSequence()
			if err != nil {
				return err
			}
			cid := strconv.FormatUint(id, 10)

			j, err := json.Marshal(apiReq)
			if err != nil {
				return err
			}

			err = b.Put([]byte(cid), j)
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
}
