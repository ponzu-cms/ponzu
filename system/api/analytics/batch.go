package analytics

import (
	"encoding/json"
	"strconv"

	"github.com/boltdb/bolt"
)

// batchInsert is effectively a specialized version of SetContentMulti from the
// db package, iterating over a []apiRequest instead of []url.Values
func batchInsert(batch []apiRequest) error {
	err := store.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("requests"))
		if err != nil {
			return err
		}

		for _, apiReq := range batch {
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
