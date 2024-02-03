package recovery_key

import (
	"github.com/boltdb/bolt"
)

// SetRecoveryKey generates and saves a random secret key to verify an email
// address submitted in order to recover/reset an account password
func (repo *repository) SetRecoveryKey(email, key string) error {
	return repo.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}

		err = b.Put([]byte(email), []byte(key))
		if err != nil {
			return err
		}

		return nil
	})
}
