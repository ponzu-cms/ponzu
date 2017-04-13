package analytics

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

// batchInsert is effectively a specialized version of SetContentMulti from the
// db package, iterating over a []apiRequest instead of []url.Values
func batchInsert(requests chan apiRequest) error {
	var reqs []apiRequest
	batchSize := len(requestChan)

	for i := 0; i < batchSize; i++ {
		reqs = append(reqs, <-requestChan)
	}

	err := store.Update(func(tx *bolt.Tx) error {
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

// batchPrune takes a duration to evaluate apiRequest dates against. If any of
// the apiRequest timestamps are before the threshold, they are removed.
// TODO: add feature to alternatively backup old analytics to cloud
func batchPrune(threshold time.Duration) error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	max := today.Add(threshold)

	// iterate through all request data
	err := store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__requests"))

		err := b.ForEach(func(k, v []byte) error {
			var r apiRequest
			err := json.Unmarshal(v, &r)
			if err != nil {
				return err
			}

			// delete if timestamp is below or equal to max
			ts := time.Unix(r.Timestamp/1000, 0)
			if ts.Equal(max) || ts.Before(max) {
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
