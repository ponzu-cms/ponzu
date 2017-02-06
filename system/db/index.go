package db

import (
	"bytes"
	"encoding/json"

	"github.com/boltdb/bolt"
)

// Index gets the value from the namespace at the key provided
func Index(namespace, key string) ([]byte, error) {
	val := &bytes.Buffer{}
	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(index(namespace)))
		if b == nil {
			return nil
		}

		v := b.Get([]byte(key))

		_, err := val.Write(v)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return val.Bytes(), nil
}

// SetIndex sets a key/value pair within the namespace provided and will return
// an error if it fails
func SetIndex(namespace, key string, value interface{}) error {
	return store.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(index(namespace)))
		if err != nil {
			return err
		}

		val, err := json.Marshal(value)
		if err != nil {
			return err
		}

		return b.Put([]byte(key), val)
	})
}

// DeleteIndex removes the key and value from the namespace provided and will
// return an error if it fails. It will return nil if there was no key/value in
// the index to delete.
func DeleteIndex(namespace, key string) error {
	return store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(index(namespace)))
		if b == nil {
			return nil
		}

		return b.Delete([]byte(key))
	})
}

// DropIndex removes the index and all key/value pairs in the namespace index
func DropIndex(namespace string) error {
	return store.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte(index(namespace)))
		if err == bolt.ErrBucketNotFound {
			return nil
		}

		if err != nil {
			return err
		}

		return nil
	})
}

func index(namespace string) string {
	return "__index_" + namespace
}
