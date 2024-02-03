package root

import (
	"github.com/boltdb/bolt"
	"strconv"
	"strings"
)

func (repo *repository) NextIDSequence(namespace string) (string, error) {
	var specifier string // i.e. __pending, __sorted, etc.
	if strings.Contains(namespace, "__") {
		spec := strings.Split(namespace, "__")
		namespace = spec[0]
		specifier = "__" + spec[1]
	}

	var cid string
	err := repo.db.View(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(namespace + specifier))
		if err != nil {
			return err
		}

		// get the next available ID and convert to string
		// also set effectedID to int of ID
		id, err := b.NextSequence()
		if err != nil {
			return err
		}

		cid = strconv.FormatUint(id, 10)
		return nil
	})

	if err != nil {
		return "", err
	}

	return cid, nil
}

// IsValidID checks that an ID from a DB target is valid.
// ID should be an integer greater than 0.
// ID of -1 is special for new posts, not updates.
// IDs start at 1 for auto-incrementing
func (repo *repository) IsValidID(id string) bool {
	if i, err := strconv.Atoi(id); err != nil || i < 1 {
		return false
	}

	return true
}
