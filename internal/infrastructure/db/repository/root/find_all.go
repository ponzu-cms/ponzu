package root

import (
	"bytes"
	"github.com/boltdb/bolt"
)

// FindAll retrieves all items from the database within the provided namespace
func (repo *repository) FindAll(entityType string) ([]interface{}, error) {
	var posts []interface{}
	err := repo.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(entityType))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		numKeys := b.Stats().KeyN
		posts = make([]interface{}, 0, numKeys)

		return b.ForEach(func(k, v []byte) error {
			post, err := repo.MarshalEntity(entityType, bytes.NewBuffer(v))
			if err != nil {
				return err
			}

			posts = append(posts, post)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return posts, nil
}
