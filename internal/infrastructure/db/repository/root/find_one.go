package root

import (
	"bytes"
	"github.com/boltdb/bolt"
	"log"
	"strings"
)

// FindOneByTarget retrieves one item from the database. Non-existent values will return an empty []byte
// The `target` argument is a string made up of namespace:id (string:int)
func (repo *repository) FindOneByTarget(target string) (interface{}, error) {
	t := strings.Split(target, ":")
	ns, id := t[0], t[1]

	val := &bytes.Buffer{}
	err := repo.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		_, err := val.Write(b.Get([]byte(id)))
		if err != nil {
			log.Println(err)
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return repo.MarshalEntity(target, val)
}
