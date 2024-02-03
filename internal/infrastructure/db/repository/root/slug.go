package root

import (
	"fmt"
	"github.com/boltdb/bolt"
)

func (repo *repository) UniqueSlug(slug string) (string, error) {
	// check for existing slug in __contentIndex
	err := repo.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(contentIndexName))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		original := slug
		exists := true
		i := 0
		for exists {
			s := b.Get([]byte(slug))
			if s == nil {
				exists = false
				return nil
			}

			i++
			slug = fmt.Sprintf("%s-%d", original, i)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return slug, nil
}
