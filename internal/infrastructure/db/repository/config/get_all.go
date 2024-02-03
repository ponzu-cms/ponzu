package config

import (
	"bytes"
	"fmt"
	"github.com/boltdb/bolt"
)

// GetAll returns all configuration from the db
func (repo *repository) GetAll() ([]byte, error) {
	val := &bytes.Buffer{}
	err := repo.store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__config"))
		if b == nil {
			return fmt.Errorf("error finding bucket: %s", "__config")
		}

		_, err := val.Write(b.Get([]byte("settings")))
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
