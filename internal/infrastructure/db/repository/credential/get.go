package recovery_key

import (
	"bytes"
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/fanky5g/ponzu/internal/domain/entities"
)

func (repo *repository) GetByUserId(userId string, credentialType entities.CredentialType) (*entities.CredentialHash, error) {
	val := &bytes.Buffer{}
	err := repo.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		_, err := val.Write(b.Get(repo.getKey(userId, credentialType)))
		return err
	})

	if err != nil {
		return nil, err
	}

	if val.Len() == 0 {
		return nil, nil
	}

	var credentialHash entities.CredentialHash
	if err = json.NewDecoder(val).Decode(&credentialHash); err != nil {
		return nil, err
	}

	return &credentialHash, nil
}
