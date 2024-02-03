package recovery_key

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/fanky5g/ponzu/internal/domain/entities"
)

func (repo *repository) SetCredential(hash *entities.CredentialHash) error {
	return repo.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}

		value := &bytes.Buffer{}
		if err = json.NewEncoder(value).Encode(hash); err != nil {
			return fmt.Errorf("failed to encode credential hash: %v", err)
		}

		return b.Put(repo.getKey(hash.UserId, hash.Type), value.Bytes())
	})
}
