package recovery_key

import (
	"bytes"
	"github.com/boltdb/bolt"
)

// GetRecoveryKey gets a previously set recovery key to verify an email address
// submitted in order to recover/reset an account password
func (repo *repository) GetRecoveryKey(email string) (string, error) {
	key := &bytes.Buffer{}

	err := repo.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		_, err := key.Write(b.Get([]byte(email)))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return key.String(), nil
}
