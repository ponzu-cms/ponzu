package recovery_key

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
)

type repository struct {
	db *bolt.DB
}

var bucketName = "__recoveryKeys"

func New(db *bolt.DB) (interfaces.RecoveryKeyRepositoryInterface, error) {
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	}); err != nil {
		return nil, fmt.Errorf("failed to create storage bucket: %v", bucketName)
	}

	return &repository{db: db}, nil
}
